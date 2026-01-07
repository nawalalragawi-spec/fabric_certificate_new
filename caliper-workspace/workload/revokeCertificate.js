'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class RevokeCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        
        // استخدام نفس نمط المعرف (ID) الذي تم استخدامه في مرحلة الإصدار (Issue)
        // ملاحظة: يجب أن تكون هذه الشهادات قد أُنشئت بالفعل في مرحلة سابقة من الاختبار
        const certID = `CERT_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            // استدعاء دالة الإلغاء بدلاً من الحذف الفيزيائي
            contractFunction: 'RevokeCertificate', 
            contractArguments: [certID],
            readOnly: false
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new RevokeCertificateWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
