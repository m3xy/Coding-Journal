/**
 * Submission.jsx
 * Author: 190019931
 * 
 * Page for displaying a submission file
 * 
 */

import React, { useEffect, useState } from 'react'
import { Card, Col, Container, ListGroup, Row } from 'react-bootstrap';
import { useParams } from 'react-router-dom'
import Explorer from './Explorer'
import Code from "./Code"
import axiosInstance from "../Web/axiosInstance";

const submissionEndpoint = '/submission'

function Submission() {  
    const params = useParams(); //URL parameters (Submission ID)

    const [file, setFile] = useState({ID:null, submissionId:null, path:"", name:""});
    const [submission, setSubmission] = useState({
        ID:null,
        name:"",
        license:"",
        files:[file],
        authors:[],
        reviewers:[],
        categories:[],
        metaData:{
            abstract:"",
            reviews:[]
        }
    });

    useEffect(() => {

        //Get submission via its ID
        axiosInstance.get(submissionEndpoint + "/" + params.id)
            .then((response) => {
                console.log(response)
                //Get first file of submission
                // axiosInstance.get(fileEndpoint + "/" + response.data.files[0].ID)
                //     .then((response) => {

                //         //Set the 
                //         setFile(response.data);
                //         setCode(atob(response.data.base64Value));
                //     }).catch((error) => {
                //         console.log(error);
                //     })

                setSubmission(response.data);
            }).catch((error) => {
                console.log(error);
            })
    }, [])


    const filesHTML = "Accordion?";

    return (
        <Container>
            <br />
            <Row>
                <Col>
                <Card border="light" className='row no-gutters'>
                <Card.Header><b>Submission</b></Card.Header>
                <Card.Body>
                    <Card.Title>{submission.name}</Card.Title>
                    <Card.Text>{submission.metaData.abstract}</Card.Text>
                    <ListGroup variant="flush">
                    
                        <ListGroup.Item>Authors</ListGroup.Item>
                        <ListGroup.Item>Reviewers</ListGroup.Item>
                        <ListGroup.Item>Reviews</ListGroup.Item>
                    </ListGroup>
                    
                </Card.Body>
                <Card.Footer></Card.Footer>
                </Card>
                </Col>

            </Row>
            <Row>
                <Col xs={3}>
                    <Explorer />
                <p>(component changes state with setFile())</p>
                </Col>
                <Col>
                    <Code id={submission.files[0].ID} />
                </Col>
            </Row>
        </Container>
    )
}

export default Submission
