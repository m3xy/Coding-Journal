import React from "react";
import { Link, withRouter } from "react-router-dom";


function navButton(props) {
}

function Navigation(props) {
	return (
		<div className="navigation">
			<nav class="navbar navbar-expand navbar-dark bg-dark">
				<div class="container">
					<Link class="navbar-brand" to="/">
						React Multi-Page Website
					</Link>
					<div>
						<ul class="navbar-nav ml-auto">
							<li
								class="nav-item"
							>
								<Link class="nav-link" to="/">
									Home
									<span class="sr-only">(current)</span>
								</Link>
							</li>
							<li
								class="nav-item"
							>
								<Link class="nav-link" to="/about">
									About
									<span class="sr-only">(current)</span>
								</Link>
							</li>
							<li
								class="nav-item"
							>
								<Link class="nav-link" to="/contact">
									Contact	
									<span class="sr-only">(current)</span>
								</Link>
							</li>
							<li
								class="nav-item"
							>
								<Link class="nav-link" to="/login">
									Login
									<span class="sr-only">(current)</span>
								</Link>
							</li>
							<li
								class="nav-item"
							>
								<Link class="nav-link" to="/register">
									Register
									<span class="sr-only">(current)</span>
								</Link>
							</li>
						</ul>
					</div>
				</div>
			</nav>
		</div>
	)
}

export default Navigation;
