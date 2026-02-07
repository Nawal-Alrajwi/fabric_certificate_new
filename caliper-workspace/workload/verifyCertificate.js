'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // تأكد أن نمط التسمية (Cert_) يطابق ما استخدمته في دالة الإصدار
        const certID = `Cert_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            // التعديل هنا: يجب أن يكون ReadCertificate
            contractFunction: 'ReadCertificate', 
            contractArguments: [certID],
            readOnly: true // عمليات القراءة والتحقق هي Read-only
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new VerifyCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
