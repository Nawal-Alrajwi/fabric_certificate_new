'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    /**
    * إرسال معاملة إصدار شهادة متوافقة مع العقد الذكي المطور.
    */
    async submitTransaction() {
        this.txIndex++;
        
        // توليد معرف فريد لكل شهادة بناءً على رقم العامل وفهرس المعاملة
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;
        
        // البيانات التي سيتم إرسالها لتخزينها وحساب الـ SHA-3 الـخاص بها
        const request = {
            contractId: 'basic',
            contractFunction: 'IssueCertificate',
            contractArguments: [
                certID,                               // id
                'Student ' + this.txIndex,             // studentName
                'Computer Science',                    // major
                'University of Sanaa',                 // university
                '2025-12-28',                          // issueDate
                'Excellent',                           // grade
                'Admin_01'                             // issuerID
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
