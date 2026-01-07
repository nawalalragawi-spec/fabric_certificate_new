'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto');

class IssueCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.workerIndex = -1;
        this.totalWorkers = -1;
    }

    /**
    * تهيئة المتغيرات الأساسية للمختبر (Worker)
    */
    async initializeWorkloadModule(workerIndex, totalWorkers, numberProtocols, workloadContext) {
        this.workerIndex = workerIndex;
        this.totalWorkers = totalWorkers;
    }

    /**
    * الدالة الأساسية التي تنفذ المعاملة (إصدار شهادة)
    */
    async submitTransaction() {
        // 1. توليد معرف فريد للشهادة بناءً على وقت التنفيذ ورقم العامل
        const certId = `Cert_${this.workerIndex}_${Date.now()}`;
        
        // 2. محاكاة بيانات الشهادة
        const owner = 'Student_' + Math.floor(Math.random() * 1000);
        const issuer = 'University_Authority';
        const issueDate = new Date().toISOString();
        
        // 3. توليد بصمة SHA-256 عشوائية (لمحاكاة بصمة ملف PDF)
        const randomContent = crypto.randomBytes(20).toString('hex');
        const certHash = crypto.createHash('sha256').update(randomContent).digest('hex');

        // 4. إعداد الوسائط (Arguments) لإرسالها للعقد الذكي (Go)
        // الترتيب يجب أن يطابق تعريف الدالة في Go: 
        // IssueCertificate(ctx, id, owner, issuer, issueDate, certHash)
        const requestSettings = {
            contractId: 'basic', // تأكد أن هذا هو اسم الـ Chaincode المسجل في الشبكة
            contractFunction: 'IssueCertificate',
            contractArguments: [certId, owner, issuer, issueDate, certHash],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(requestSettings);
    }

    async cleanupWorkloadModule() {
        // دالة تنظيف إذا لزم الأمر بعد انتهاء الاختبار
    }
}

/**
 * تصدير الدالة ليتمكن Caliper من تشغيلها
 */
function createWorkloadModule() {
    return new IssueCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
