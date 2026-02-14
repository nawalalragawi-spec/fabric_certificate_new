'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class UnlockCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    /**
     * تهيئة البيانات قبل بدء الاختبار
     */
    async initializeWorkloadModule(workerIndex, totalWorkers, numberofIndices, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, numberofIndices, sutAdapter, sutContext);
    }

    async submitTransaction() {
        this.txIndex++;
        
        // استخدام نفس نمط المعرف (ID) الذي تم استخدامه عند الإصدار
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        /**
         * بناءً على ورقة عمر سعد، عملية الفتح تأخذ خيارات:
         * option 1: فتح البيانات العامة فقط
         * option 2: فتح البيانات الخاصة
         * option 3: فتح الهاش للتحقق
         * سنقوم هنا بإرسال الخيار رقم 2 كمثال للمطابقة مع البحث
         */
        const unlockOption = 2; 

        const request = {
            contractId: 'basic',            // اسم العقد الذكي المنصب في شبكتك
            contractFunction: 'UnlockCertificate', 
            contractArguments: [
                certID, 
                unlockOption.toString()     // إرسال الخيار كـ string ليتوافق مع استقبال العقد
            ],
            readOnly: false                 // هذه عملية تحديث للحالة (Write Transaction)
        };

        try {
            await this.sutAdapter.sendRequests(request);
        } catch (error) {
            console.error(`فشل عملية فتح الشهادة ${certID}: ${error.message}`);
        }
    }

    async cleanupWorkloadModule() {
        // العمليات التنظيفية إن وجدت
    }
}

function createWorkloadModule() {
    return new UnlockCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
