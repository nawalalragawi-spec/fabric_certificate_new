'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class RevokeCertificateWorkload extends WorkloadModuleBase {
    constructor() {
        super();
        this.txIndex = 0;
    }

    async submitTransaction() {
        this.txIndex++;
        // نحذف نفس المعرف الذي تم استخدامه في مرحلة الإصدار لضمان وجوده
        const certID = `cert_${this.workerIndex}_${this.txIndex}`;

        const request = {
            contractId: 'basic',
            // التعديل: تم تغيير اسم الدالة لتطابق الموجود في عقد Go (RevokeCertificate)
            contractFunction: 'RevokeCertificate', 
            // الدالة تتوقع وسيطاً واحداً فقط وهو المعرف (id)
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