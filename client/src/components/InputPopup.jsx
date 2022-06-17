import {
	FormControl,
	Input,
	FormHelperText,
	InputLabel,
	Button,
	Grid,
} from "@mui/material";
import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { searchMovie } from "../actions/movieActions";
import Movies from "./Movies";

const InputPopup = ({order}) => {
	const dispatch = useDispatch();
	const [searchResults, setSearchResults] = useState([{}]);
	const [query, setQuery] = useState("");
	const moviesState = useSelector((state) => state.searchMovieReducer);
	const { loading, error, movies } = moviesState;
	const API_KEY = "api_key=f5b09f41dd618287a9b5c448e9ef8d72";
	const BASE_URL = "https://api.themoviedb.org/3";
	const API_URL = BASE_URL + "/search/movie?" + API_KEY + "&query=";

	const handleSubmit = (e) => {
		e.preventDefault();
		console.log("submitted");
		dispatch(searchMovie(API_URL + query));
		// console.log(movies);
		setSearchResults(movies);
		// console.log(searchResults);
	};

	const handleChange = (e) => {
		// console.log(query);
		setQuery(e.target.value);
	};

	return (
		<div
			style={{
				display: "flex",
				flexDirection: "column",
				alignItems: "center",
				padding: "1rem",
				justifyContent: "flex-start",
				width: "80%",
				height: "80%",
				color: "black",
				backgroundColor: "white",
				borderRadius: "10px",
				overflowY: "scroll",
			}}
		>
			{/* <Grid container> */}
			<form onSubmit={handleSubmit}>
				<FormControl>
					<InputLabel htmlFor="my-input" color={"primary"}>
						Enter Movie:
					</InputLabel>
					<Input
						id="my-input"
						aria-describedby="my-helper-text"
						onChange={handleChange}
					/>
					<Button variant="contained" color="primary" onClick={handleSubmit}>
						Search
					</Button>
				</FormControl>
			</form>
			<div
				style={{
					display: "flex",
					flexDirection: "column",
					alignItems: "flex-start",
					padding: "1rem",
					justifyContent: "flex-start",
					width: "80%",
					height: "80%",
					color: "black",
					backgroundColor: "white",
					borderRadius: "10px",
					overflowY: "scroll",
				}}
			>
				<h1 style={{ alignSelf: "center" }}>Search Results</h1>
				{loading && <h1>Loading</h1>}
				{error && <h1>Error ocured</h1>}

				{movies &&
					movies.map((movie, i) => {
						return (
							// <Grid item>
                            <Movies key={i} movie={movie} serial={i} order={order}/>
							// </Grid>
						);
					})}
			</div>
			{/* </Grid> */}
		</div>
	);
};

export default InputPopup;
