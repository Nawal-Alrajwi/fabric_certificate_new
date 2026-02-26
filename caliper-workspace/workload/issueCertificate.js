'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;
        const studentName = `Student_${this.workerIndex}_${this.txIndex}`;
        const degree = 'Bachelor of Computer Science';
        const issuer = 'Digital University';
        const certHash = crypto.createHash('sha256').update(certID + studentName).digest('hex'); 
        const issueDate = new Date().toISOString().split('T')[0]; 

        const request = {
            contractId: 'basic', 
            contractFunction: 'IssueCertificate', 
            // ✅ تم تعديل الترتيب هنا ليطابق توقيع دالة Go (التاريخ أولاً ثم الهاش)
            contractArguments: [certID, studentName, degree, issuer, issueDate, certHash],
            readOnly: false
        };

        return this.sutAdapter.sendRequests(request); 
    }
}

module.exports = { createWorkloadModule: () => new IssueCertificateWorkload() };
