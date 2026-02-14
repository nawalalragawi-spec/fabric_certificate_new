'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class LockCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    /**
    * تهيئة الجولة: التأكد من أن المتغيرات جاهزة
    */
    async initializeWorkloadModule(workerIndex, totalWorkers, numberofIndices, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, numberofIndices, sutAdapter, sutContext);
    }

    async submitTransaction() {
        this.txIndex++;
        
        // استخدام نفس نمط المعرف (ID) المستخدم في دورة الإصدار لضمان استهداف شهادات موجودة
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        /**
         * إعداد طلب المعاملة لدالة LockCertificate
         */
        const request = {
            contractId: 'basic',
            contractFunction: 'LockCertificate',
            contractArguments: [certID],
            readOnly: false // هذه معاملة تحديث (Invoke) لتغيير حالة الشهادة
        };

        try {
            // إضافة await هنا تضمن أن السكربت سينتظر نتيجة الشبكة
            // إذا رفض العقد الذكي المعاملة (مثلاً المستدعي ليس المالك)، سينتقل الكود إلى بلوك catch
            await this.sutAdapter.sendRequests(request);
        } catch (error) {
            // تسجيل الخطأ في التيرمينال للمساعدة في التصحيح أثناء الاختبار
            // Caliper سيسجل هذه المعاملة كفشل (Failed Transaction) تلقائياً
            console.error(`خطأ تقني في قفل الشهادة ${certID}: ${error.message}`);
        }
    }

    async cleanupWorkloadModule() {
        // عمليات التنظيف عند انتهاء الاختبار
    }
}

function createWorkloadModule() {
    return new LockCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
