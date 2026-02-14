'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class QueryAllWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    /**
    * في بيئة "عمر سعد"، الاستعلام عن كافة الشهادات يعطي فكرة عن 
    * أداء الشبكة عند قراءة السجل العام (Public Ledger).
    */
    async submitTransaction() {
        const request = {
            contractId: 'basic',
            // مطابقة اسم الدالة المحدثة في العقد الذكي
            contractFunction: 'GetAllCertificates', 
            // الدالة لا تأخذ وسائط لأنها تقرأ كل شيء من السجل العام
            contractArguments: [],
            // تحديد أنها قراءة فقط لضمان عدم استهلاك Resources المعالجة (Endorsement)
            readOnly: true
        };

        try {
            await this.sutAdapter.sendRequests(request);
        } catch (error) {
            console.error(`فشل الاستعلام عن الشهادات: ${error.message}`);
        }
    }
}

function createWorkloadModule() {
    return new QueryAllWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
