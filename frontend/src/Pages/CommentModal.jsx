/**
 * CommentModal.jsx
 * Author: 190010714
 * 
 * A Modal component dedicated to displaying comments.
 */
import React from 'react'
// import {Helmet} from "react-helmet";
import  axiosInstance  from "../Web/axiosInstance";
import {Modal, Button, InputGroup, FormControl} from "react-bootstrap";

const commentEndpoint = 'submission/file/newcomment' ;


class CommentModal extends React.Component{
  constructor(props) {
    super(props);
    this.state = {
      show: false,
      val: "Will it work"
    };
    this.showModal = this.showModal.bind(this);
    this.hideModal = this.hideModal.bind(this);
  }

  showModal = () => {
    this.setState({ show: true });
  }

  hideModal = () => {
    this.setState({ show: false });
  }

  onSubmit = () => {
    this.hideModal
    console.log(this.state.val);
  }

  /**
  * Author: 190010714
  * Sends a POST request to the go server to uplaod a new comment
  *
  * @param file the file ID for the file on which the comment was made
  * @param submission the submission ID for the submission in which the file is in
  * @param author the author of the comment
  * @param content the content of the comment
  */
  uploadComment(file, submission, author, content) {
    let data = {
        filePath: file,
        submissionId: submission,
        author: author,
        content: content
    };
    axiosInstance.post(commentEndpoint, data)
                 .then(() => {
                   console.log("Received: " + files);
                 })
                 .catch((error) => {
                   console.log(error)
                 })
  }


  // TODO - Remove the hard coding.
  componentDidMount() {
    // You can call the Prism.js API here
    setTimeout(() => Prism.highlightAll(), 0)
    console.log(window.submission);

    let userID = null;                          //Preparing to get userID from session cookie
    let cookies = document.cookie.split(';');   //Split all cookies into key value pairs
    for(let i = 0; i < cookies.length; i++){    //For each cookie,
        let cookie = cookies[i].split("=");     //  Split key value pairs into key and value
        if(cookie[0].trim() == "userID"){       //  If userID key exists, extract the userID value
            userID = cookie[1].trim();
            break;
        }
    }

    if(userID === null){                        //If user has not logged in, disallow submit
        console.log("Not logged in");
        return;
    }

    this.uploadComment('CountToFifteen.java', 8, userID.substring(11,50), this.state.val);
    console.log("Comment submitted");
  }


  render() {
    return (
      <>
      <Button variant="primary" onClick={this.showModal}>
        Reviewer Comments
      </Button>

      <Modal show={this.state.show} >
        <Modal.Header closeButton onClick={this.hideModal}>
          <Modal.Title>Comments</Modal.Title>
        </Modal.Header>
        <Modal.Body style={{textAlign: 'center',}}>Enter Reviewer Comments Below
            <InputGroup>
            {/* <InputGroup.Text>Enter Reviewer Comment</InputGroup.Text> */}
            <FormControl as="textarea" aria-label="With textarea" value={this.state.val} onChange={e => this.setState({ val: e.target.value })}
          type="text" />
          </InputGroup>
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={this.hideModal}>
            Close
          </Button>
          <Button variant="primary" onClick={this.onSubmit} >
            Save Changes
          </Button>
        </Modal.Footer>
      </Modal>
    </>
    )
  }
}

export default CommentModal;
