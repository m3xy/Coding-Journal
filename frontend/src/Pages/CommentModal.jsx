import React from 'react'
import {Helmet} from "react-helmet";
import {Modal, Button} from "react-bootstrap";



class CommentModal extends React.Component{

  constructor() {
    super();
    this.state = {
      show: false
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
        <Modal.Body>Woohoo, you're reading this text in a modal!</Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={this.hideModal}>
            Close
          </Button>
          <Button variant="primary" onClick={this.hideModal}>
            Save Changes
          </Button>
        </Modal.Footer>
      </Modal>
    </>
    )
  }

}

export default CommentModal;