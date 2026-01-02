'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyDiplomaWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // استخدام معرف موجود تم إصداره مسبقاً في جولة الـ Issue
        const certID = 'DIP_TEST'; 

        const request = {
            contractId: 'diploma',
            // التعديل الجوهري: استدعاء دالة التحقق الأمنية وليس مجرد القراءة
            contractFunction: 'VerifyDiploma', 
            contractArguments: [certID],
            readOnly: true 
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new VerifyDiplomaWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
