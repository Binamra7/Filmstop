import "./App.css";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Home from "./pages/Home";
import About from "./pages/About";
import RateMovies from "./pages/RateMovies";
import Navbar from "./components/Navbar";

function App() {
	return (
		<div className="App">
			<BrowserRouter>
				<Navbar />
				<Routes>
					<Route path="/" element={<Home />} />
					<Route path="/about" element={<About />} />
					<Route path="/rate" element={<RateMovies />} />
				</Routes>
			</BrowserRouter>
		</div>
	);
}

export default App;
