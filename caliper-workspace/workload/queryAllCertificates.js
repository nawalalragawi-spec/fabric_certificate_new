'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    async submitTransaction() {
        const request = {
            contractId: 'basic',
            // التعديل: تم تغيير اسم الدالة لتطابق الموجود في عقد Go الذكي
            contractFunction: 'GetAllCertificates', 
            contractArguments: [],
            readOnly: true
        };

        await this.sutAdapter.sendRequests(request);
    }
}

function createWorkloadModule() {
    return new QueryAllWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;