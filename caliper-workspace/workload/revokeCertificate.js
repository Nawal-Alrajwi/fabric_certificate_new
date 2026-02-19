'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class RevokeCertificateWorkload extends WorkloadModuleBase {

    constructor() {
        super();
        this.txIndex = 0;
        this.maxIssuedPerWorker = 1000; // عدد تقريبي تم إنشاؤه في Round Issue
    }

    async submitTransaction() {

        this.txIndex++;

        const workerId = this.workerIndex || 0;

        // اختيار ID موجود فعليًا داخل النطاق
        const randomIndex =
            (this.txIndex % this.maxIssuedPerWorker) + 1;

        const certID = `CERT_${workerId}_${randomIndex}`;

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
