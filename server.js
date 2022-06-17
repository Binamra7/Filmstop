const express = require("express");
const app = express();
const cors = require("cors");
const dotenv = require("dotenv");
const mongoose = require("mongoose");

//use CORS middleware
app.use(cors());
//use dotenv
dotenv.config();

app.use(express.json());

const port = process.env.PORT || 5000;

//connect to mongoose
const uri = process.env.MONGO_URI;
mongoose.connect(uri, { useNewUrlParser: true, useUnifiedTopology: true });

const connection = mongoose.connection;
connection.once("open", () => {
	console.log("MongoDB database connection established successfully");
});

const userRoute = require("./routes/userRoute");

app.use("/api/movies", userRoute);

app.listen(port, () => {
	console.log(`Server started on port ${port}`);
});
