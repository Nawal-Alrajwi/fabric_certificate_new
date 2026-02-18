'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class RevokeCertificateWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {

        this.txIndex++;

        // يجب أن تكون الشهادة موجودة من مرحلة Issue
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;

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
