const mongoose = require("mongoose")
const patientModel = require("../models/patient")

exports.registerPatient = async(req, res) => {
    try{

        const  {name, patient_id, medicine_taken, slot } = req.body;
    
        if (!name || !patient_id || medicine_taken===undefined || !slot) {
            return res.status(400).json({message:"Data missing"});
        }
    
    
        const newUser = new patientModel({
            name,
            patient_id,
            medicine_taken,
            slot
        });
    
        await newUser.save();
    
        return res.status(200).json({message:"User created successfully"});
    }catch(error) {
        console.error(error)
    }
}

exports.getPatient = async(req, res) => {
    try {

        const {patientid} = req.params;
        if (!patientid) {
            return res.status(400).json({message:"Id missing in params"});
        }

        const patientData = await patientModel.aggregate([
            {
                $match: {
                    patient_id : Number(patientid)
                }
            },
            {
                $group:{
                    _id: "$patient_id",
                    name: {$first: "$name"},
                    data :{
                        $push:{
                            slot:"$slot",
                            time:"$time",
                            medicine_taken: "$medicine_taken"
                        }
                    }
                }
            },
            {
                $project: {
                    _id: 0,
                    patient_id: "$_id",
                    name: 1,
                    data: 1
                }
            }
        ]);
    
        return res.status(200).json(patientData);
    }catch(error) {
        console.error(error);
    }
}

exports.getAllpatients = async (req, res) => {

    try{
        const users = await patientModel.find()

        return res.status(200).json(users)
    }catch(error) {
        console.error(error);
    }
    

}

exports.getProbability = async (req, res) => {
    try {
        const { patient_id } = req.params;

        if (!patient_id) {
            return res.status(400).json({ message: "Patient ID is required" });
        }

        const records = await patientModel.find({ patient_id: patient_id });

        if (records.length === 0) {
            return res.status(404).json({ message: "No records found for this patient" });
        }

        const totalCounts = { 1: 0, 2: 0, 3: 0 };
        const missedCounts = { 1: 0, 2: 0, 3: 0 };

        records.forEach((entry) => {
            const slot = entry.slot;
            totalCounts[slot]++;
            if (!entry.medicine_taken) {
                missedCounts[slot]++;
            }
        });

        const probabilities = {
            Slot1: totalCounts[1] ? Math.round((missedCounts[1] / totalCounts[1]) * 100) : 0,
            Slot2: totalCounts[2] ? Math.round((missedCounts[2] / totalCounts[2]) * 100) : 0,
            Slot3: totalCounts[3] ? Math.round((missedCounts[3] / totalCounts[3]) * 100) : 0,
        };

        return res.status(200).json(probabilities);

    } catch (error) {
        console.error("Error in getProbability:", error);
        return res.status(500).json({ message: "Internal server error" });
    }
};
