const mongoose = require('mongoose');

const patientSchema = new mongoose.Schema({
    patient_id:{
        type:Number,
        required:true
    },
    name:{
        type:String,
        required:true
    },
    time: {
        type: Date,
        default: Date.now
    },
    medicine_taken:{
        type:Boolean,
        required:true
    },
    slot:{
        type:Number,
        required:true,
        enum: [1, 2, 3]
    }
});

const patient = mongoose.model("patient", patientSchema)
module.exports = patient