import React from "react";
import { Link } from "react-router-dom";
import "./Navbar.css";
import { Button } from "@mui/material";
import FilmStop from "../assets/FilmStop.JPG";
import { useAuth0 } from "@auth0/auth0-react";
const Navbar = () => {
	const { user, isAuthenticated, loginWithRedirect, loginWithPopup, logout } =
		useAuth0();
	if (isAuthenticated) {
		// console.log(user);
		// console.log("loggedin");
	}
	return (
		<div className="navbar">
			<div className="navbar-logo">
				FilmStop
				{/* <img src={FilmStop} alt="FilmStop" /> */}
			</div>
			<div className="navbar-links">
				<div className="navbar-link">
					<Link to="/">Home</Link>
				</div>
				<div className="navbar-link">
					<Link to="/about">About</Link>
				</div>
				<div className="navbar-link">
					<Link to="/rate">Rate Movies</Link>
				</div>
			</div>
			<div>
				<div className="user">
					{isAuthenticated ? (
						<div className="logout">
							<img
								style={{ margin: "0px" }}
								className="profile-pic"
								src={user.picture}
								alt="Profile"
							/>
							<p style={{ margin: "0 5px" }}>{user.given_name}</p>
							<Button
								style={{ margin: "0 10px" }}
								variant="contained"
								color="primary"
								onClick={() => logout()}
							>
								Logout
							</Button>
						</div>
					) : (
						<Button
							variant="contained"
							onClick={() => {
								loginWithPopup();
							}}
						>
							Login
						</Button>
					)}
				</div>
			</div>
		</div>
	);
};

export default Navbar;
