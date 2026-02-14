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
        
        // استخدام نفس نمط المعرف (ID) المستخدم في دورة الإصدار لضمان استهداف شهادات موجودة فعلاً
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        /**
         * دالة LockCertificate في العقد الذكي تتوقع وسيط واحد وهو ID الشهادة.
         * العقد الذكي سيتحقق داخلياً أن المستدعي (Invoker) هو المالك.
         */
        const request = {
            contractId: 'basic',
            contractFunction: 'LockCertificate',
            contractArguments: [certID],
            readOnly: false // هذه معاملة كتابة لأنها تغير حالة الشهادة في السجل (Ledger)
        };

        try {
            await this.sutAdapter.sendRequests(request);
        } catch (error) {
            console.error(`فشل عملية قفل الشهادة ${certID}: ${error.message}`);
        }
    }

    async cleanupWorkloadModule() {
        // عمليات التنظيف إن وجدت
    }
}

function createWorkloadModule() {
    return new LockCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
