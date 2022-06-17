import React, { useEffect, useState } from "react";
import MovieCardInput from "../components/MovieCardInput";
import { useSelector, useDispatch } from "react-redux";
import { Button } from "@mui/material";
import { fetchUserMovies, submitMovies } from "../actions/movieActions";
import { useAuth0 } from "@auth0/auth0-react";

const FiveMovies = () => {
	const dispatch = useDispatch();
	const moviesState = useSelector((state) => state.selectMovieReducer);
	// const [movies, setMovies] = useState([]);
	let movies = moviesState.movies;
	const [hasSubmitted, setHasSubmitted] = useState(false);
	const userMoviesState = useSelector((state) => state.fetchUserMoviesReducer);

	const { user } = useAuth0();
	useEffect(() => {
		dispatch(fetchUserMovies(user));
		const userMovies = userMoviesState.movies;
		if (userMovies.length > 0) {
			console.log("this");
			setHasSubmitted(true);
			// movies = [...userMovies];
			console.log("movies", movies);
		}
	}, []);
	if (hasSubmitted) {
		const userMovies = userMoviesState.movies;
		movies = [...userMovies];
	}
	const isValid = () => {
		let count = 0;
		movies.forEach((movie) => {
			if (typeof movie === "undefined") {
				return false;
			} else count++;
		});
		return count === 5;
	};

	const submitMoviesHandler = (e) => {
		if (isValid) {
			e.preventDefault();
			dispatch(submitMovies(movies || [], user || {}));
		}
	};

	return (
		<>
			{hasSubmitted && <h1>Thank you for submitting...</h1>}
			<div
				style={{
					display: "flex",
					justifyContent: "center",
					alignItems: "center",
					marginTop: "50px",
				}}
			>
				<MovieCardInput movie={movies[0]} order="1st" />
			</div>
			<div
				style={{
					display: "flex",
					flexWrap: "wrap",
					justifyContent: "space-evenly",
					alignItems: "center",
					marginTop: "4rem",
				}}
			>
				<MovieCardInput movie={movies[1]} order="2nd" />
				<MovieCardInput movie={movies[2]} order="3rd" />
			</div>
			<div
				style={{
					display: "flex",
					flexWrap: "wrap",
					width: "100%",
					justifyContent: "space-around",
					alignItems: "center",
					marginTop: "4rem",
				}}
			>
				<MovieCardInput movie={movies[3]} order="4th" />
				<MovieCardInput movie={movies[4]} order="5th" />
			</div>
			<Button
				variant="contained"
				color="primary"
				disabled={!isValid()}
				onClick={submitMoviesHandler}
			>
				Submit
			</Button>
		</>
	);
};

export default FiveMovies;
