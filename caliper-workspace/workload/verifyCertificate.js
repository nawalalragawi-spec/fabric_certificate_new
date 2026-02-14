'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class VerifyCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        // استخدام نفس نمط المعرف (ID) الذي استخدمناه في IssueCertificate
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        // الهاش الأصلي (Plaintext) الذي نريد التحقق منه 
        // ملاحظة: يجب أن يكون نفس الهاش الذي تم تشفيره وإرساله في مرحلة الإصدار
        const originalHash = "hash_data_" + this.txIndex;

        /**
         * منطق البحث (عمر سعد):
         * دالة VerifyCertificate في العقد الذكي ستتحقق أولاً:
         * if cert.IsLocked { return error }
         * لذا يجب التأكد في ملف benchConfig.yaml من تشغيل دورة Unlock قبل هذه الدورة.
         */

        const request = {
            contractId: 'basic',
            contractFunction: 'VerifyCertificate', 
            contractArguments: [
                certID, 
                originalHash // نرسل الهاش الصريح ليقوم العقد بفك تشفير المخزن ومقارنته به
            ],
            readOnly: true // عملية التحقق قراءة فقط حسب منطق Hyperledger Fabric
        };

        try {
            await this.sutAdapter.sendRequests(request);
        } catch (error) {
            // إذا كانت الشهادة مقفلة، كليبر سيسجل فشل المعاملة Transaction Failure
            // وهذا بحد ذاته نتيجة بحثية تثبت كفاءة نظام الخصوصية لديك
            console.error(`فشل التحقق من الشهادة ${certID}: ${error.message}`);
        }
    }
}

function createWorkloadModule() {
    return new VerifyCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
