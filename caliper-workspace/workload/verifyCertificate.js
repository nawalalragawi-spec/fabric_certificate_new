'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');
const crypto = require('crypto');

class VerifyCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.workerIndex = -1;
    }

    async initializeWorkloadModule(workerIndex, totalWorkers, numberProtocols, workloadContext) {
        this.workerIndex = workerIndex;
    }

    /**
    * الدالة التي تحاكي عملية التحقق من الشهادة
    */
    async submitTransaction() {
        // 1. تحديد معرف الشهادة المراد فحصها
        // (يفضل أن يكون مطابقاً لمعرف تم إنشاؤه في ملف issueCertificate)
        const certId = `Cert_${this.workerIndex}_${Math.floor(Math.random() * 100)}`;
        
        // 2. محاكاة عملية حساب Hash لملف شهادة يرفعه المستخدم للتحقق
        // في الواقع، هذا الـ Hash ينتج عن الملف المرفوع حالياً
        const randomContent = 'Actual_Certificate_Content_To_Verify'; 
        const currentFileHash = crypto.createHash('sha256').update(randomContent).digest('hex');

        // 3. إعداد الطلب لدالة التحقق في العقد الذكي (Go)
        const requestSettings = {
            contractId: 'basic',
            contractFunction: 'VerifyCertificate',
            contractArguments: [certId, currentFileHash],
            readOnly: true // التحقق هو عملية قراءة واستعلام فقط
        };

        // إرسال الطلب واستلام النتيجة (true إذا كانت سليمة، false إذا تم التلاعب بالـ Hash)
        await this.sutAdapter.sendRequests(requestSettings);
    }
}

function createWorkloadModule() {
    return new VerifyCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
