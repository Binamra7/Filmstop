const mongoose = require("mongoose");

const Schema = mongoose.Schema;

const movieSchema = new Schema({
	id: {
		type: String,
		required: true,
	},
	title: {
		type: String,
		required: true,
	},
	poster_path: {
		type: String,
		required: true,
	},
	overview: {
		type: String,
		trim: true,
	},
});

const userSchema = new Schema({
	username: {
		type: String,
		required: true,
		trim: true,
	},
	userId: {
		type: String,
		required: true,
		trim: true,
	},
	profile_pic: {
		type: String,
		required: true,
	},
	movies: [movieSchema],
});

const User = mongoose.model("User", userSchema);
module.exports = User;
