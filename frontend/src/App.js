/**
 * App.js
 * 
 * This file contains the over-arching structure and functionality of the web page.
 * To add a new "page", add a Route line here. To add a link to that file to the main
 * navigation bar, see ./components/Navigation.jsx and add a <Link> there
 */

import logo from './logo.svg';
import './App.css';
import {BrowserRouter as Router, Route, Switch} from "react-router-dom";
import { Navigation, Footer, Home, About, Contact, Login} from "./components";

function App() {
  return (
    <div className="App">
		<Router>
			<Navigation />
			<Switch>
				<Route path="/" exact component={() => <Home />} />
				<Route path="/about" exact component={() => <About />} />
				<Route path="/contact" exact component={() => <Contact />} />
				<Route path="/login" exact component={Login} />
			</Switch>
			<Footer />
		</Router>
    </div>
  );
}

export default App;
