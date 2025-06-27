const express = require("express");
require('dotenv').config();
const cors = require("cors")
const connectDB = require("../db/dbConnection")


const connectionString = process.env.MONGO_URI

const patient = require("../routes/patientRoute")

connectDB(connectionString)
const app = express();
app.use(cors({origin:'*'}))
app.use(express.json());

const PORT = process.env.PORT || 4000

app.get("/", (req, res)=> {
    res.json("Backend Running")
})

app.use("/patient", patient)

app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
});
