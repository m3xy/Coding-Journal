import React, { useState, useEffect } from "react";
import FolderTree, { testData } from 'react-folder-tree';
import { useNavigate } from "react-router-dom";
import 'react-folder-tree/dist/style.css';
import {Row, Tab, ListGroup, Col, } from "react-bootstrap";
import axiosInstance from "../Web/axiosInstance";

const profileEndpoint = '/users';

function getUserID() {
  let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
  for(let i = 0; i < cookies.length; i++){    //For each cookie,
      let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
      if(cookie[0].trim() == "userId") {       //  If userId key exists, extract the userId value
          return cookie[1].trim();
      }
  }
  return null;
}

function Submissions() {
  const navigate = useNavigate()
  const [firstname, setFirstname] = useState('Manuel')
  const [lastname, setLastname] = useState('Hinke')
  const [usertype, setUsertype] = useState(0)
  const [email, setEmail] = useState('')
  const [phonenumber, setPhoneNumber] = useState('000000')
  const [organization, setOrganization] = useState('None')
  const [userSubmissions, setSubmissions] = useState('')


  function openSubmission(submissionsID) {
    navigate("/code/" + submissionsID)
  }

  if (getUserID() === null) {
    navigate("/login")
  }

  useEffect(() => {
    axiosInstance.get(profileEndpoint + "/" + getUserID())
      .then((response) => {
        console.log(response.data);
        setFirstname(response.data.firstname)
        setLastname(response.data.lastname)
        setUsertype(response.data.usertype)
        setEmail(response.data.email)
        setPhoneNumber(response.data.phonenumber)
        setOrganization(response.data.organization)
        setSubmissions(response.data.submissions)
      })
      .catch(() => {
        return (<div></div>)
      })
  }, [])
  //Get user comments
  const fakeSubmission = [["0", "Arrays and Matrixes"], ["1", "Computational Power of a System"], ["2", "Discrete Math"], ["3", "Turing Proof"], ["4", "On the Complexity of the Machine"], ["5", "Database Mergers"], ["6", "List Aquistions"] ]
  const fakeFiles = [["0", "test.java"], ["1", "src"], ["2", "images"], ["3", ".env"], ["4", "README.md"], ["5", "run.sh"], ["6", "Dockerbuild"] ]


  const comments = []
  const userTypes = ["None", "Publisher", "Reviewer", "Reviewer-Publisher", "User"]
  const submissions = fakeSubmission.map(([id, name]) => {
    return (
      <ListGroup.Item as = "li" key={id} action href={id}>
          {name}
        </ListGroup.Item>
    );
  })
  const files = fakeFiles.map(([id, name]) => {
    if(name.indexOf('.') == -1){
      return (
        <ListGroup.Item as = "li" key={id} action href={id}>
            {name}
          </ListGroup.Item>
      );
    }
    else{
      return (
        <ListGroup.Item as = "li" key={id} action href={id} disabled>
            {name}
          </ListGroup.Item>
      );
    }
    
  })

  
  const treeState = {
    name: 'Submission',
    checked: 0.5,   // half check: some children are checked
    isOpen: false,   // this folder is opened, we can see it's children
    children: [
      {
        name: 'README.md', checked: 0
      },
      {
        name: 'src',
        checked: 0,
        isOpen: false,
        children:
          [
            { name: 'App.jsx', checked: 0 },
            { name: 'Link.jsx', checked: 1 },
          ],
      },
    ],
  };



const onTreeStateChange = (state, event) => console.log(state, event);

return (
  <div className="col-md-6 offset-md-3" style={{ textAlign: "left"}}>
    <br />
    <h2>{firstname + " " + lastname}</h2>
    <label>({userTypes[usertype]})</label>
    <br /><br />
    <Tab.Container id="list-group-tabs-example" defaultActiveKey="0" >
  <Row>
    <Col sm={4}>
      {submissions.length > 0 ? (
						<ListGroup>{submissions}</ListGroup>
					) : (
						<div className="text-center" style={{color:"grey"}}><i>No posts</i></div>
					)
					}
    </Col>
    <Col sm={8}>
      <Tab.Content>
        <Tab.Pane eventKey="0">
        {submissions.length > 0 ? (
						<ListGroup>{files}</ListGroup>
					) : (
						<div className="text-center" style={{color:"grey"}}><i>No posts</i></div>
					)
					}
        </Tab.Pane>
        <Tab.Pane eventKey="1">
          <FolderTree
            data={treeState}
            onChange={onTreeStateChange}
            showCheckbox={false}
          />
        </Tab.Pane>
        <Tab.Pane eventKey="2">
          <FolderTree
            data={treeState}
            onChange={onTreeStateChange}
            showCheckbox={false}
          />
        </Tab.Pane>
        <Tab.Pane eventKey="3">
          <FolderTree
            data={treeState}
            onChange={onTreeStateChange}
            showCheckbox={false}
          />
        </Tab.Pane>
        <Tab.Pane eventKey="4">
          <FolderTree
            data={treeState}
            onChange={onTreeStateChange}
            showCheckbox={false}
          />
        </Tab.Pane>
        <Tab.Pane eventKey="5">
          <FolderTree
            data={treeState}
            onChange={onTreeStateChange}
            showCheckbox={false}
          />
        </Tab.Pane>
        <Tab.Pane eventKey="6">
          <FolderTree
            data={treeState}
            onChange={onTreeStateChange}
            showCheckbox={false}
          />
        </Tab.Pane>
      </Tab.Content>
    </Col>
  </Row>
</Tab.Container>

  </div>
)
}
export default Submissions;