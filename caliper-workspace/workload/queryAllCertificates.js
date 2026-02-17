'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllCertificatesWorkload extends WorkloadModuleBase {

    async submitTransaction() {

        const request = {
            contractId: 'basic',
            contractFunction: 'QueryAllCertificatesWithPagination',
            contractArguments: ['100', ''],   // pageSize=100 , bookmark=''
            readOnly: true
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new QueryAllCertificatesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
