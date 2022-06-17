const router = require("express").Router();
const User = require("../models/users.model");
const Rank = require("../models/rank.model");

router.get("/", (_, res) => {
	let topFive = [];
	Rank.find()
		.then((movies) => {
			movies.sort((a, b) => {
				return b.score - a.score;
			});
			topFive = movies.slice(0, 5);
			res.json(topFive);
		})
		.catch((err) => {
			res.status(400).json(err);
		});
});

router.post("/user", (req, res) => {
	const { user } = req.body;
	// console.log(user);
	User.find({ userId: user?.sub || "" }, (err, user) => {
		if (err) {
			res.status(400).json(err);
		} else {
			if (user.length > 0) {
				// console.log("user found", user);
				res.json(user[0].movies);
			} else {
				res.status(404).json({ message: "User not found" });
			}
		}
	});
});

router.post("/user/add", (req, res) => {
	const { formattedMovies, user } = req.body;

	const newMovie = new User({
		movies: formattedMovies,
		username: user.given_name,
		userId: user.sub,
		profile_pic: user.picture,
	});
	// console.log("executing post route");
	newMovie
		.save()
		.then(() => res.status(200).json("Movie added!"))
		.catch((err) => res.status(400).json("Error: " + err));
});

router.post("/rank/add", (req, res) => {
	// console.log("executing rank route");

	const { formattedMovies, user } = req.body;
	let moviesRank = [];
	const ranker = () => {
		for (let i = 0; i < 5; i++) {
			Rank.find({ id: formattedMovies[i].id }, (err, movies) => {
				if (movies.length > 0) {
					// console.log("found");
					Rank.findByIdAndUpdate(
						movies[0]._id,
						{
							id: movies[0].id,
							title: movies[0].title,
							poster_path: movies[0].poster_path,
							overview: movies[0].overview,
							score: movies[0].score + (100 - i * 20),
						},
						(err, newScore) => {
							if (err) console.log(err);
							// console.log("new score: ", newScore);
						}
					);
				}
				if (movies.length === 0) {
					// console.log("not found");
					moviesRank.push({
						id: formattedMovies[i].id,
						title: formattedMovies[i].title,
						poster_path: formattedMovies[i].poster_path,
						overview: formattedMovies[i].overview,
						score: 100 - i * 20,
					});
				}
			});
		}
	};
	ranker();
	setTimeout(() => {
		Rank.insertMany(moviesRank);
		res.status(200).json("Rank Added");
	}, 4000);
});

module.exports = router;
