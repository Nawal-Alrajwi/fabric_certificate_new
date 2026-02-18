'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllCertificatesWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.bookmarks = {}; // تخزين bookmark لكل Worker
    }

    async submitTransaction() {

        const workerId = this.workerIndex || 0;
        const currentBookmark = this.bookmarks[workerId] || '';

        const request = {
            contractId: 'basic',
            contractFunction: 'QueryAllCertificatesWithPagination',
            contractArguments: ['50', currentBookmark], // pageSize=50
            readOnly: true
        };

        const response = await this.sutAdapter.sendRequests(request);

        // تحديث bookmark من النتيجة القادمة من chaincode
        if (response && response.length > 0) {
            try {
                const payload = JSON.parse(response[0].result);

                if (payload.nextBookmark) {
                    this.bookmarks[workerId] = payload.nextBookmark;
                } else {
                    // إذا انتهت الصفحات نبدأ من جديد
                    this.bookmarks[workerId] = '';
                }

            } catch (err) {
                // في حال فشل parsing نعيد bookmark
                this.bookmarks[workerId] = '';
            }
        }
    }
}

function createWorkloadModule() {
    return new QueryAllCertificatesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
