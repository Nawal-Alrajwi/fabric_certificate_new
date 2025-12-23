'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyDiplomaWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // التعديل هنا: البحث عن المعرف الثابت المضمن في كل دفعة إصدار
        const certID = 'DIP_TEST'; 

        const request = {
            contractId: 'diploma',
            contractFunction: 'ReadDiploma', 
            contractArguments: [certID],
            readOnly: true // معاملات القراءة يتم ضبطها كـ true لتحسين الأداء
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new VerifyDiplomaWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
