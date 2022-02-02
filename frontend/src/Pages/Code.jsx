/**
 * Code.jsx
 * Author: 190019931
 * 
 * This file stores the info for rendering the Code page of our Journal
 */

import React, {useState, useEffect, useRef} from "react";
import {useParams} from 'react-router-dom';
import axiosInstance from "../Web/axiosInstance";
import {Container, Row, Col, Form, InputGroup, Card, Breadcrumb, Modal, Button, Toast} from "react-bootstrap"
import MonacoEditor from 'react-monaco-editor';

const codeEndpoint = 'submission/file'
const commentEndpoint = 'submission/file/newcomment';

function Code(props) {

    const [code, setCode] = useState('// type your code...');
    const [submissionName, setSubmissionName] = useState('Submission');
    const [submissionId, setSubmissionId] = useState(0);
    const [filePath, setFilePath] = useState('App/Pages/index.js');

    // const [fileName, setFileName] = useState('App/Pages/index.js');

    // let { submissionId, filePath } = useParams();

    const [comments, setComments] = useState({
        1:[ {submissionId: null, filePath: null, author: "John Doe", base64Value: "Looks Good!"}, 
            {submissionId: null, filePath: null, author: "Jane Doe", base64Value: "I disagree."},
            {submissionId: null, filePath: null, author: "Jim Doe", base64Value: "I have 500 more citations than both of you, I can assure you, this code is mediocre."}
          ]
    }        
    );
    const [show, setShow] = useState(false);

    const monacoRef = useRef();
    const [theme, setTheme] = useState('vs');
    const [language, setLanguage] = useState('javascript');
    const [lineNumber, setLineNumber] = useState(1);
    const [isLoading, setLoading] = useState(false);


    useEffect(() => {
        // setCode('//Request Submission ID: ' + window.submissionId);
        if(typeof window.submission !== 'undefined'){
            setCode(atob(window.submission.split(",")[1]));
            setSubmissionId(window.submissionId);
            setFilePath(window.submissionName);
        }

        // if (isLoading) {
        //     simulateNetworkRequest().then(() => {
        //       setLoading(false);
        //     });
        // }

        //Remember to also request comments for submission
        // axiosInstance.post(codeEndpoint, null, {params: {submissionId, filePath}})
        //             .then((response) => {
        //                 console.log(response);
        //             })
        //             .catch((error) => {
        //                 console.log(error)
        //             });
    })

    const editorDidMount = (editor, monaco) => {
        console.log('editorDidMount', editor);

        // monaco.languages.registerHoverProvider(language, {
        //     provideHover: function(model, position) {
        //         // console.log(model.getWordAtPosition(position).word); //Able to retrieve word
        //         // console.log(position); //Can also support any arbritrary range within code (Comment on lines/words)

        //         return {
        //             range: new monaco.Range(position.lineNumber, 1, model.getLineMaxColumn(position.lineNumber)),
        //             contents: [
        //                 { value: '**Comments**' },
        //                 { supportHtml: true, value: "[Reviewer Comments](http://localhost:23409/comment)"}
        //             ]
        //         }
        //     }
        // });

        editor.addAction({

            id: 'Comment',                                                  // An unique identifier of the contributed action.
            label: 'Comment',                                               // A label of the action that will be presented to the user. (Right-click)
            keybindings: [ monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyM ],   // An optional array of keybindings for the action.
            precondition: null,                                             // A precondition for this action.
            keybindingContext: null,                                        // A rule to evaluate on top of the precondition in order to dispatch the keybindings.
            contextMenuGroupId: 'navigation',
            contextMenuOrder: 1.5,
        
            // Method that will be executed when the action is triggered.
            // @param editor The editor instance is passed in as a convenience
            run: function (ed) {
                // let comment = prompt("Comment on line " + ed.getPosition().lineNumber, "Type Here");
                setLineNumber(ed.getPosition().lineNumber);
                handleShow();
            }
        });
        // editor.deltaDecorations(
        //     [],
        //     [
        //         {
        //             range: new monaco.Range(1, 1, 1, 1),
        //             options: { 
        //                 isWholeLine: true,
        //                 linesDecorationsClassName: 'myLineDecoration',
        //                 inlineClassName: 'myInlineDecoration',
        //                 hoverMessage: [{value: "Hello"}, {value: "[link](#comments)"}],
        //                 glyphMarginHoverMessage: [{value: "Bye"}, {value: "[link](https://www.google.com)"}],
        //                 glyphMarginClassName: 'myGlyphMarginClass'
        //             }
        //         }
        //     ]
        // )
        editor.focus();
    };

    const onChange = (newValue, e) => {
        console.log('onChange', newValue, e);
        setCode(newValue)
    };

    const options = {
        selectOnLineNumbers: true,
        glyphMargin: true
        // readOnly: true
    };

    const postComment = (e) => {
        e.preventDefault();
        let comment = document.getElementById("CommentText").value;

        if (comment == null || comment == "") {
            console.log("No comment written");
        } else {

            let userId = null;                          //Preparing to get userId from session cookie
            let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
            for(let i = 0; i < cookies.length; i++){    //For each cookie,
                let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
                if(cookie[0].trim() == "userId"){       //  If userId key exists, extract the userId value
                    userId = cookie[1].trim();
                    break;
                }
            }
            if(userId === null){                        //If user has not logged in, disallow submit
                console.log("Not logged in");
                return;
            }

            let data = {
                submissionId: submissionId,
                filePath: filePath,
                author: userId,
                base64Value: btoa(comment)
            }
            console.log(data);
            axiosInstance.post(commentEndpoint, data)
                        .then((response) => {
                            console.log(response);
                            document.getElementById("CommentText").value = "";
                        })
                        .catch((error) => {
                            console.log(error);
                        });
        }
        
    }

    const handleClose = () => {
        setShow(false);
    }

    const handleShow = () => {
        setShow(true);
    }

    const handleClick = () => setLoading(true);

    // const commentsHTML = Object.values(comments).map((line, i) => {
    //     return line.map((comment, j) => {
    //         return (
    //             <Toast className="d-inline-block m-1" key={i.toString() + ":" + j.toString()}>
    //                 <Toast.Header closeButton={false}>
    //                     {/* <img src="holder.js/20x20?text=%20" className="rounded me-2" alt="" /> */}
    //                     <strong className="me-auto">{comment.author}</strong>
    //                     <small>{"Line: " + line}</small>
    //                 </Toast.Header>
    //                 <Toast.Body>{comment.base64Value}</Toast.Body>
    //             </Toast>
    //         );
    //     })
    // })

    const commentsHTML = Object.entries(comments).map((line, i) => {
        return line[1].map((comment, j) => {
            return (
                <Toast className="d-inline-block m-1" key={i + ":" + j}>
                    <Toast.Header closeButton={false}>
                        {/* <img src="holder.js/20x20?text=%20" className="rounded me-2" alt="" /> */}
                        <strong className="me-auto">{comment.author}</strong>
                        <small>{"Line: " + line[0]}</small>
                    </Toast.Header>
                    <Toast.Body>{comment.base64Value}</Toast.Body>
                </Toast>
            );
        })
    })

    const filePathHTML = filePath.split("/").map((part, i) => {
        return (<Breadcrumb.Item href={'/code/' + submissionId + '/' + filePath.match(new RegExp('^(.*?)' + part))[0]} key={i}>{part}</Breadcrumb.Item>);
    })

    return(
        <Container>
            <br />
            <Card>
            <Card.Header>{submissionName}</Card.Header>
            <Card.Body>
            <Card.Title>
                <Breadcrumb>
                    {filePathHTML}
                </Breadcrumb>
            </Card.Title>
            <Card.Text>Abstract</Card.Text>
            <Row>
                <Col>
                    <InputGroup size="sm" className="mb-3">
                        <InputGroup.Text id="inputGroup-sizing-sm">Language: </InputGroup.Text>
                        <Form.Select defaultValue={language} size="sm" onChange={(e) => { setLanguage(e.target.value) }}>
                            <option value="javascript">Javascript</option>
                            <option value="html">HTML</option>
                            <option value="css">CSS</option>
                            <option value="json">JSON</option>
                            <option value="java">Java</option>
                            <option value="python">Python</option>
                        </Form.Select>
                    </InputGroup>
                </Col>
                <Col>
                    <InputGroup size="sm" className="mb-3">
                        <InputGroup.Text id="inputGroup-sizing-sm">Theme: </InputGroup.Text>
                        <Form.Select size="sm" onChange={(e) => { setTheme(e.target.value) }}>
                            <option value="vs">Visual Studio</option>
                            <option value="vs-dark">Visual Studio Dark</option>
                            <option value="hc-black">High Contrast Dark</option>
                        </Form.Select>
                    </InputGroup>
                </Col>
            </Row>
            <Row>
            <Modal show={show} onHide={handleClose} size="lg">
            <Form onSubmit={postComment}>
                <Modal.Header closeButton>
                    <Modal.Title>Comments</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    {commentsHTML}
                    <div className="d-grid gap-2">
                        <Button variant="link" disabled={isLoading} onClick={!isLoading ? handleClick : null}>
                            {isLoading ? 'Loading…' : 'Load more'}
                        </Button>
                    </div>
                    <br />
                    
                        <Form.Group className="mb-3" controlId="CommentText">
                            <Form.Label>Enter a comment below:</Form.Label>
                            <Form.Control as="textarea" rows={3} aria-describedby="lineNumber"/>
                            <Form.Text id="lineNumber" muted>(Line: {lineNumber})</Form.Text>
                        </Form.Group>              
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Close
                    </Button>
                    <Button variant="primary" type="submit">
                        Post comment
                    </Button>
                </Modal.Footer>
            </Form>
            </Modal>
            </Row>
            <Row>
                <Col>
                    <MonacoEditor
                        ref={monacoRef}
                        height="1000"
                        language={language}
                        theme={theme}
                        value={code}
                        options={options}
                        onChange={onChange}
                        editorDidMount={editorDidMount}
                    />
                </Col>
            </Row>
            </Card.Body>
            <Card.Footer className="text-muted">2 days ago</Card.Footer>
            </Card>
        </Container>
    )
}

export default Code;