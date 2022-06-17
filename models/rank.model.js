const mongoose = require("mongoose");
const Schema = mongoose.Schema;

const rankSchema = new Schema({
	id: {
		type: String,
		required: true,
		trim: true,
	},
	title: {
		type: String,
		required: true,
		trim: true,
	},
	poster_path: {
		type: String,
		required: true,
		trim: true,
	},
	overview: {
		type: String,
		trim: true,
	},
	score: {
		type: Number,
		required: true,
		trim: true,
	},
});

const Rank = mongoose.model("Rank", rankSchema);

module.exports = Rank;
