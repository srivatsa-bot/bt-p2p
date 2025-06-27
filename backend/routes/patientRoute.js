const  express = require("express")

const router = express.Router();

const {registerPatient, getPatient, getAllpatients, getProbability} = require("../controllers/patientController")


router.post("/", registerPatient )
router.get("/:patientid", getPatient)
router.get("/",getAllpatients)
router.get("/probability/:patient_id", getProbability);

module.exports = router