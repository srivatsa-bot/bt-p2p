const mongoose = require ("mongoose")

const connectDB = async (dbConnectionString)=>{
    try{
        const connection = await mongoose.connect(dbConnectionString);
        console.log("Connection Successfull")
        console.log(`PORT: ${connection.connection.port}`);
        console.log(`NAME: ${connection.connection.name}\n`);
    }catch(error){
        console.error(error)
    }
}

module.exports = connectDB;