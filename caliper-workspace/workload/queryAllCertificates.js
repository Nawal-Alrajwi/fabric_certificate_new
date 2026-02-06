'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllCertificatesWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        // تخزين الـ bookmark للاستعلامات المقسمة
        this.bookmarks = {};
    }

    async submitTransaction() {
        // الطريقة الأولى: استعلام محسّن بـ pagination
        // بدلاً من جلب كل الشهادات، نجلبها على دفعات
        const workerID = this.workerIndex || 0;
        const bookmark = this.bookmarks[workerID] || '';

        const request = {
            contractId: 'basic',
            contractFunction: 'QueryAllCertificatesWithPagination',
            contractArguments: ['100', bookmark],
            readOnly: true
        };

        try {
            const result = await this.sutAdapter.sendRequests(request);
            
            // تحديث الـ bookmark للطلب التالي
            if (result && result.length > 0) {
                try {
                    const parsed = JSON.parse(result[0].result);
                    if (parsed.nextBookmark) {
                        this.bookmarks[workerID] = parsed.nextBookmark;
                    }
                } catch (e) {
                    // تجاهل أخطاء التحليل وإعادة تعيين الـ bookmark
                    delete this.bookmarks[workerID];
                }
            }
        } catch (error) {
            // في حالة الخطأ، أعد تعيين الـ bookmark
            delete this.bookmarks[workerID];
            throw error;
        }
    }
}

function createWorkloadModule() {
    return new QueryAllCertificatesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
