import { Input, InputLabel } from "@mui/material";
import axios from "axios";
import React, { useState } from "react";
import FiveMovies from "../components/FiveMovies";

const RateMovies = () => {
	const Movie = {
		id: -1,
		title: "",
		poster_path: "",
		release_date: "",
		overview: "",
	};
	const [movies, setMovies] = useState(new Array(5).fill(Movie));

	// const API_KEY = "api_key=f5b09f41dd618287a9b5c448e9ef8d72";
	// const BASE_URL = "https://api.themoviedb.org/3";
	// const API_URL = BASE_URL + "/search/movie?" + API_KEY + "&query=";
	// console.log(API_URL);
	// axios.get(API_URL + "avengers").then((res) => {
	// 	console.log(res.data.results[0]);
	// 	// setMovies(res.data.results);
	// });
	return (
		<div>
			<FiveMovies movies={movies} />
		</div>
	);
};

export default RateMovies;
