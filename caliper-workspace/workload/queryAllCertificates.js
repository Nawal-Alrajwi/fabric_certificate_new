'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllCertificatesWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    async submitTransaction() {
        const request = {
            contractId: 'basic', 
            // يجب أن يتطابق الاسم مع الدالة في كود Go (التي أسميناها GetAllCertificates)
            contractFunction: 'GetAllCertificates', 
            contractArguments: [],
            readOnly: true // العمليات التي تقرأ فقط من قاعدة البيانات تكون true
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new QueryAllCertificatesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
