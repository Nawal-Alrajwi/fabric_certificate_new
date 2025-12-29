'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    /**
    * عملية التحقق المطورة باستخدام المقارنة المباشرة للبصمة الرقمية (Hash).
    */
    async submitTransaction() {
        this.txIndex++;
        
        // معرف الشهادة المستهدف
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;
        
        // البيانات التي سيتم إعادة حساب الـ SHA-3 لها في العقد الذكي لمطابقتها
        // ملاحظة: يجب أن تتطابق هذه البيانات تماماً مع ما تم إرساله في دالة Issue
        const providedData = `${certID}Student ${this.txIndex}University of Sanaa2025-12-28`;

        const request = {
            contractId: 'basic',
            // نستخدم الدالة الجديدة التي أضفناها في الـ Smart Contract
            contractFunction: 'VerifyCertificate', 
            contractArguments: [certID, providedData],
            readOnly: true // التحقق عملية قراءة فقط لتسريع الاستجابة
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new VerifyCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
