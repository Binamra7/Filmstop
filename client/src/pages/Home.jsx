import React, { useEffect } from "react";
import FiveMoviesHome from "../components/FiveMoviesHome";
import { useDispatch, useSelector } from "react-redux";
import { getMovies } from "../actions/movieActions";

const Home = () => {
	const dispatch = useDispatch();
	const moviesState = useSelector((state) => state.getMoviesReducer);
	const { loading, error, movies } = moviesState;
	useEffect(() => {
		document.title = "Home";
		dispatch(getMovies());
	}, [dispatch]);
	return (
		<div>
			<h1>Highest Rated Movied By Audience of FilmStop</h1>
			{loading && <h1>Loading</h1>}
			{error && <h1>Error ocured</h1>}
			{movies && movies.length > 0 && <FiveMoviesHome movies={movies} />}
		</div>
	);
};

export default Home;
