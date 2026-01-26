'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto');

class RevokeCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, numberofIndices, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, numberofIndices, sutAdapter, sutContext);
        this.txIndex = 0; // تصفير العداد عند بداية الجولة
    }

    async submitTransaction() {
        this.txIndex++;
        
        // --- المنطق المضمون لنجاح العملية ---
        // نحن نعلم أن الجولة الأولى (Issue) أصدرت شهادات تبدأ من العداد 1.
        // لذا سنطلب من كل عامل حذف الشهادات التي بدأ بصناعتها من الرقم 1 صعوداً.
        // هذا يضمن أن "المعرف" موجود فعلياً في الـ Ledger.
        const rawID = `Cert_${this.workerIndex}_${this.txIndex}`;

        // تشفير المعرف باستخدام SHA-3 (مطابق تماماً لما يتوقعه العقد الذكي)
        const certID = crypto.createHash('sha3-256').update(rawID).digest('hex');

        const requestSettings = {
            contractId: 'diploma', 
            contractFunction: 'RevokeCertificate', 
            contractArguments: [certID, 'Administrative decision for revocation'],
            readOnly: false
        };

        try {
            await this.sutAdapter.sendRequests(requestSettings);
        } catch (error) {
            // طباعة تفاصيل الخطأ في حال حدوثه (مثل Asset not found)
            console.error(`Worker ${this.workerIndex} Error: ${error.message}`);
        }
    }
}

function createWorkloadModule() {
    return new RevokeCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
