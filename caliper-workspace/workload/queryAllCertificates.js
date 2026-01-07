'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllCertificatesWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    /**
    * الدالة الأساسية التي تنفذ عملية الاستعلام
    */
    async submitTransaction() {
        // إعداد إعدادات الطلب
        // نستخدم 'GetAllCertificates' وهي الدالة التي أضفناها في عقد Go
        const requestSettings = {
            contractId: 'basic', 
            contractFunction: 'GetAllCertificates',
            contractArguments: [], // هذه الدالة لا تحتاج بارامترات في العقد الذكي
            readOnly: true // تحديد أنها عملية قراءة فقط لتحسين الأداء في الاختبار
        };

        await this.sutAdapter.sendRequests(requestSettings);
    }
}

/**
 * تصدير الدالة ليتمكن Caliper من تشغيلها
 */
function createWorkloadModule() {
    return new QueryAllCertificatesWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
