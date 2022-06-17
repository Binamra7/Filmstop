import React, { useState } from "react";
import { Typography } from "@mui/material";
import { useDispatch } from "react-redux";
import { selectMovieAction } from "../actions/movieActions";

const Movies = (props) => {
	const dispatch = useDispatch();

	const IMG_URL = "https://image.tmdb.org/t/p/w300";
	const source = IMG_URL + props.movie.poster_path;
	const handleClick = () => {
		const order = parseInt(props.order.substring(0, 1)) - 1;
		dispatch(selectMovieAction(props.movie, order));
	};

	return (
		<>
			<div
				onClick={handleClick}
				style={{
					display: "flex",
					margin: "10px 0",
					border: "1px solid grey",
					width: "100%",
				}}
			>
				<img src={source} alt="Failed to load" style={{ width: "200px" }} />
				<div
					style={{ display: "flex", flexDirection: "column", margin: "10px" }}
				>
					<h1>
						{props.serial + 1}.&nbsp;
						{props.movie?.original_title}
					</h1>
					<Typography>
						<i>{props.movie?.release_date.substring(0, 4)}</i>
					</Typography>
					<p>{props.movie?.overview}</p>
				</div>
			</div>
		</>
	);
};

export default Movies;
