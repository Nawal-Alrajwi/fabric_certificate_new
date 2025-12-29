'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class RevokeCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    /**
    * إرسال طلب إلغاء شهادة وحذف بصمتها الرقمية من السجل.
    */
    async submitTransaction() {
        this.txIndex++;
        
        // استخدام نفس المعرف الذي تم إنشاؤه في مرحلة الإصدار لضمان نجاح الحذف
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            // تأكد أن الاسم يطابق الدالة في العقد الذكي المطور (DeleteAsset هي الافتراضية في Fabric)
            contractFunction: 'DeleteAsset', 
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
