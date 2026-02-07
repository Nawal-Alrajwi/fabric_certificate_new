'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class RevokeCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // ملاحظة هامة: يجب أن يكون certID مطابقاً للشهادات التي تم إصدارها في مرحلة الـ Create
        const certID = `Cert_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            contractFunction: 'RevokeCertificate', // التغيير هنا ليطابق دالة الإلغاء في العقد
            contractArguments: [certID],           // الدالة تتوقع فقط الـ ID
            readOnly: false                        // هذه عملية تحديث (Write) لذا تكون false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new RevokeCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
