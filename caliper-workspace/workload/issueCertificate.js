'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        // إنشاء معرف فريد للشهادة
        const certID = `Cert_${this.workerIndex}_${this.txIndex}`;
        
        // محاكاة بيانات الشهادة بناءً على هيكل الـ Certificate في العقد
        const request = {
            contractId: 'basic', // تأكد أن هذا الاسم يطابق اسم العقد في ملف config.yaml
            contractFunction: 'IssueCertificate', // تغيير الاسم ليطابق عقد الشهائد
            contractArguments: [
                certID,                     // ID
                'Student Name ' + this.txIndex, // StudentName
                'PhD in Computer Science',  // Degree
                'Sana\'a University',       // Issuer
                '2026-02-07',               // IssueDate
                'hash_value_' + this.txIndex // CertHash (تمثيل لبصمة الملف)
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
