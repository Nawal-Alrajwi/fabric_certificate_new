'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllCertificatesWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    /**
    * استعلام عن جميع الشهادات للتأكد من كفاءة استرجاع البيانات بعد تحديث الهيكل.
    */
    async submitTransaction() {
        const request = {
            contractId: 'basic',
            // تأكد أن الاسم يطابق الدالة المبرمجة في العقد الذكي (عادة GetAllAssets أو اسم مخصص)
            contractFunction: 'GetAllAssets', 
            contractArguments: [],
            readOnly: true
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new QueryAllCertificatesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
