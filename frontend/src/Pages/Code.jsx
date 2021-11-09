/**
 * Code.jsx
 * 
 * This file stores the info for rendering the Code page of our Journal
 */

import React from "react";
import Prism from "prismjs";
import CommentModal from "./CommentModal";

// import "./prism.css";

import {Helmet} from "react-helmet";

const code = `
package DataStructures.Queues;

//This program implements the concept of CircularQueue in Java
//Link to the concept: (https://en.wikipedia.org/wiki/Circular_buffer)

public class CircularQueue {
    public static void main(String[] args) {
        circularQueue cq= new circularQueue(5);
        System.out.println(cq.isEmpty());
        System.out.println(cq.isFull());
        cq.enQueue(1);
        cq.enQueue(2);
        cq.enQueue(3);
        cq.enQueue(4);
        cq.enQueue(5);

        System.out.println(cq.deQueue());
        System.out.println(cq.deQueue());
        System.out.println(cq.deQueue());
        System.out.println(cq.deQueue());
        System.out.println(cq.deQueue());
        System.out.println(cq.isFull());
        System.out.println(cq.isEmpty());
        cq.enQueue(6);
        cq.enQueue(7);
        cq.enQueue(8);
        System.out.println(cq.peek());
        System.out.println(cq.peek());
        cq.deleteQueue();

    }
}
class circularQueue{
    int[] arr;
    int topOfQueue;
    int beginningOfQueue;
    int size;
    public circularQueue(int size){
        arr=new int[size];
        topOfQueue=-1;
        beginningOfQueue=-1;
        this.size=size;
    }
    public boolean isEmpty(){
        if(beginningOfQueue==-1){
            return true;
        }else{
            return false;
        }
    }

    public boolean isFull(){
        if(topOfQueue+1==beginningOfQueue){
            return true;
        }else if(topOfQueue==size-1 && beginningOfQueue==0){
            return true;
        }else{
            return false;
        }
    }

    public void enQueue(int value){
        if(isFull()){
            System.out.println("The Queue is full!");
        }
        else if(isEmpty()) {
            beginningOfQueue=0;
            topOfQueue++;
            arr[topOfQueue]=value;
            System.out.println(value+" has been successfully inserted!");
        }else{
            if(topOfQueue+1==size){
                topOfQueue=0;
            }else{
                topOfQueue++;
            }
            arr[topOfQueue]=value;
            System.out.println(value+" has been successfully inserted!");
        }
    }

    public int deQueue(){
        if(isEmpty()){
            System.out.println("The Queue is Empty!");
            return -1;
        }else{
            int res= arr[beginningOfQueue];
            arr[beginningOfQueue]=Integer.MIN_VALUE;
            if(beginningOfQueue==topOfQueue){
                beginningOfQueue=topOfQueue=-1;
            }else if(beginningOfQueue+1==size){
                beginningOfQueue=0;
            }else{
                beginningOfQueue++;
            }
            return res;
        }

    }

    public int peek(){
        if(isEmpty()){
            System.out.println("The Queue is Empty!");
            return -1;
        }else{
            return arr[beginningOfQueue];
        }
    }

    public void deleteQueue(){
        arr=null;
        System.out.println("The Queue is deleted!");
    }

}
`.trim()

class Code extends React.Component {
  componentDidMount() {
    // You can call the Prism.js API here
    setTimeout(() => Prism.highlightAll(), 0)
  }
  render() {
    return (
    <div className="renderCode">
        <Helmet>
            <meta charSet="utf-8" />
            <title>My Title</title>
            <link rel="stylesheet" href="https://netdna.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" />
        </Helmet>
        <h2 style={{textAlign: 'center',}} >Circular Queue</h2>
        <h4 style={{textAlign: 'center',}}>Author:  </h4>
        <CommentModal/>
        <pre className="line-numbers">
            <code className="language-java">
            {code}
            </code>
        </pre>
    </div>
    )
  }
}

export default Code;