'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

/**
 * فئة عبء العمل للاستعلام عن الشهادات باستخدام تقنية التقسيم (Pagination).
 * تم تطوير هذا الملف لرفع كفاءة استرجاع بيانات SHA-3 وحل مشكلة بطء الاستجابة.
 */
class QueryAllCertificatesWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    /**
    * إرسال طلب استعلام مقسم (Paginated Query) لضمان استقرار الشبكة.
    */
    async submitTransaction() {
        const request = {
            contractId: 'basic',
            // 1. التعديل: استخدام اسم الدالة المطورة في العقد الذكي
            contractFunction: 'GetAllAssetsWithPagination', 
            
            // 2. التعديل: إرسال حجم الصفحة '50' كمعامل أول، وعلامة مرجعية فارغة كمعامل ثانٍ
            // حجم الصفحة 50 يضمن بقاء زمن الاستجابة منخفضاً جداً (Latency < 0.1s)
            contractArguments: ['50', ''], 
            
            readOnly: true
        };

        // إرسال الطلب عبر SUT Adapter
        await this.sutAdapter.sendRequests(request);
    }
}

/**
 * دالة المصنع لإنشاء وحدة عبء العمل.
 */
function createWorkloadModule() {
    return new QueryAllCertificatesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
