'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        // معرف موحد نستخدمه في جميع المراحل
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            contractFunction: 'IssueCertificate',
            contractArguments: [
                certID,                       // id
                'Student ' + this.txIndex,    // studentName
                'Computer Science',           // major (تخصص الطالب)
                'University of Sanaa',        // university (الجامعة)
                '2025-12-28',                 // issueDate (التاريخ)
                'Excellent',                  // grade (التقدير)
                'Admin_01'                    // issuerID (معرف المصدر)
            ],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new IssueCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
