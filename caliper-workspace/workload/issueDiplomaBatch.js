'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueDiplomaBatchWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.batchSize = 10; 
    }

    async submitTransaction() {
        let diplomas = [];
        
        // 1. إضافة الشهادة الثابتة لضمان نجاح جولة التحقق (Verify) بنسبة 100%
        diplomas.push({
            DiplomaID: 'DIP_TEST', // المعرف الثابت الذي سنستخدمه في ملف verifyDiploma.js
            StudentName: 'Reference Student',
            University: 'UPI University',
            Degree: 'Doctorate',
            GraduationYear: 2025
        });

        // 2. إكمال بقية الدفعة ببيانات عشوائية (نبدأ من 1 لأننا أضفنا واحدة يدوياً)
        for (let i = 1; i < this.batchSize; i++) {
            const diplomaId = 'DIP_' + Math.random().toString(36).substr(2, 9);
            diplomas.push({
                DiplomaID: diplomaId,
                StudentName: 'Student_' + diplomaId,
                University: 'University_A',
                Degree: 'Bachelor',
                GraduationYear: 2025
            });
        }

        const args = {
            contractId: 'diploma',
            contractFunction: 'CreateDiplomaBatch',
            contractArguments: [JSON.stringify(diplomas)],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(args);
    }
}

function createWorkloadModule() {
    return new IssueDiplomaBatchWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
