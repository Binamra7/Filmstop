import React, { useEffect, useState } from "react";
import MovieCardInput from "../components/MovieCardInput";
import { useSelector, useDispatch } from "react-redux";
import { Button } from "@mui/material";
import { fetchUserMovies, submitMovies } from "../actions/movieActions";
import { useAuth0 } from "@auth0/auth0-react";

const FiveMovies = () => {
	const Movie = {
		id: -1,
		title: "",
		poster_path: "",
		release_date: "",
		overview: "",
	};
	// const [movies, setMovies] = useState(new Array(5).fill(Movie));

	const dispatch = useDispatch();
	const moviesState = useSelector((state) => state.selectMovieReducer);

	let movies = moviesState.movies;

	const [hasSubmitted, setHasSubmitted] = useState(false);
	const userMoviesState = useSelector((state) => state.fetchUserMoviesReducer);

	const { isLoading, user, isAuthenticated } = useAuth0();

	// const [change, setChange] = useState(true);

	// const handleUserMovies = () => {
	// 	dispatch(fetchUserMovies());
	// 	const userMovies = userMoviesState.movies;
	// 	if (userMovies && userMovies.length > 0) {
	// 		setHasSubmitted(true);
	// 		setMovies([...userMovies]);
	// 		setChange(!change);
	// 		console.log("movies", movies);
	// 	}
	// };

	// useEffect(() => {
	// 	handleUserMovies();
	// }, [user, isAuthenticated, isLoading, dispatch]);
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
			{isLoading && <h1>Loading...</h1>}
			{!isLoading && (
				<>
					<div
						style={{
							display: "flex",
							justifyContent: "center",
							alignItems: "center",
							marginTop: "50px",
						}}
					>
						<MovieCardInput movie={movies && movies[0]} order="1st" />
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
						<MovieCardInput movie={movies && movies[1]} order="2nd" />
						<MovieCardInput movie={movies && movies[2]} order="3rd" />
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
						<MovieCardInput movie={movies && movies[3]} order="4th" />
						<MovieCardInput movie={movies && movies[4]} order="5th" />
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
			)}
		</>
	);
};

export default FiveMovies;
