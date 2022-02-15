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

import React from 'react';
import './App.css';
import { BrowserRouter as Router, Route, Routes} from 'react-router-dom';
import {Container} from "react-bootstrap"
import { Navigation, Home, Login, Register, About, Contact, Footer, Code, Upload, Profile, Comment, Submissions } from '../Pages';
 
function App() {

    return(
        <Container fluid="true">
            <Router>
                <Navigation />
                <Routes>
                    <Route path="/" element = {<Home />} />
                    <Route path="/login" element = {<Login />} />
                    <Route path="/register" element = {<Register />} />
                    <Route path="/about" element = {<About />} />
                    <Route path="/code/:submissionId/:filePath" element = {<Code />} />
                    <Route path="/code/:fileId" element = {<Code />} />
                    <Route path="/code" element = {<Code />} />
                    {/* optional URL params removed in react router v6*/}
                    <Route path="/comment" element = {<Comment />} />
                    <Route path="/contact" element = {<Contact />} />
                    <Route path="/upload" element = {<Upload />} />
                    <Route path="/profile" element = {<Profile />} />
                    <Route path="/submissions" element = {<Submissions />} />
                </Routes>
                <Footer />
            </Router>
        </Container>
    );
}

export default App;

//  class App extends React.Component {
//      constructor(props) {
//          super(props);
//          // history.listen((/*location, action*/) => {
//              // clear alert on location change
//          //     this.props.clearAlerts();
//          // });
//      }
 
//      componentDidMount() {
//          // this.deleteCookies();
//      }
 
//      /**
//       * Deletes all cookies
//       */
//      deleteCookies() {
//          var cookies = document.cookie.split(';'); 
     
//          // The "expire" attribute of every cookie is set to "Thu, 01 Jan 1970 00:00:00 GMT".
//          for (var i = 0; i < cookies.length; i++) {
//              document.cookie = cookies[i] + "=;expires=" + new Date(0).toUTCString();  //Setting all cookies expiry date to be a past date.
//          }
//      }
 
//      render() {
//          return (
//              <BrowserRouter history={history} >
//                  <Navigation />
//                  <Routes>
//                      <Route exact path="/" element = {<Home/>} />
//                      <Route path="/login" element = {<Login />} />
//                      <Route path="/register" element = {<Register />} />
//                      <Route path="/about" element = {<About />} />
//                      <Route path="/code" element = {<Code />} />
//                      <Route path="/contact" element = {<Contact/>} />
//                      <Route path="/commentModal" element = {<CommentModal />} / >
//                      <Route path="/upload" element = {<Upload/>} />
//                      <Route path="/profile" element = {<Profile/>} />
//                      {/*<Route path="*" element={<Navigate to='/' replace />} />*/}
//                  </Routes>
//                  <Footer />
//              </ BrowserRouter>
//          );
//      }
//  }