/**
 * App.jsx
 * author: 190010714, 190010425, 190019931
 *
 * This file holds the main framework for the website.
 *
 * To add a new page:
 *  (1) Code a new class/function to render a web page and add it to the Pages/ directory
 *  (2) Add a new entry in the Pages/index.js file to export your webpage and import it here
 *  (3) add a new <Route ... /> mapping a URL to your web page
 *  (4) add a new <Link /> component in the Pages/Navigation.jsx file
 */

import React from "react"
import "./App.css"
import { BrowserRouter as Router, Route, Routes } from "react-router-dom"
import { Container } from "react-bootstrap"
import {
	Login,
	Register,
	About,
	Contact,
	Footer,
	Upload,
	Profile,
	Submissions,
    Submission
} from "../Pages"
import Navigation from "../Components/Navbar"
import HomePage from "../Pages/HomePage/HomePage"

function App() {
    return(
        <Container fluid="true">
            <Router>
                <Navigation />
                <Routes>
                    <Route path="/" element = {<HomePage />} />
                    <Route path="/login" element = {<Login />} />
                    <Route path="/register" element = {<Register />} />
                    <Route path="/about" element = {<About />} />
                    <Route path="/contact" element = {<Contact />} />
                    <Route path="/upload" element = {<Upload />} />
                    <Route path="/profile" element = {<Profile />} />
                    <Route path="/submission/:id" element = {<Submission />} />
                    <Route path="/submissions" element = {<Submissions />} />
                </Routes>
                <Footer />
            </Router>
        </Container>
    );
}

export default App
