'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class IssueDiplomaBatchWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.batchSize = 10; // عدد الشهادات في الدفعة الواحدة - يمكنك رفعه لزيادة الأداء
    }

    async submitTransaction() {
        let diplomas = [];
        
        // إنشاء مجموعة (Batch) من الشهادات ببيانات عشوائية محاكية للواقع
        for (let i = 0; i < this.batchSize; i++) {
            const diplomaId = 'DIP_' + Math.random().toString(36).substr(2, 9);
            diplomas.push({
                DiplomaID: diplomaId,
                StudentName: 'Student_' + diplomaId,
                University: 'University_A',
                Degree: 'Bachelor',
                GraduationYear: 2025
            });
        }

        // تحويل المصفوفة إلى JSON string لإرسالها للـ Chaincode
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
