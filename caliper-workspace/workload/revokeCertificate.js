'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class RevokeCertificateWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {

        this.txIndex++;

        const workerId = this.workerIndex || 0;

        // استخدم نفس التسلسل الذي استخدم في Issue
        const certID = `CERT_${workerId}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            contractFunction: 'RevokeCertificate',
            contractArguments: [certID],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new RevokeCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
