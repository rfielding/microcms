import React from 'react';
//import logo from './logo.svg';
import './App.css';

import { useState, useEffect } from 'react';
import axios from 'axios';

import TreeView from '@material-ui/lab/TreeView';
import TreeItem from '@material-ui/lab/TreeItem';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';

interface Node {
  id: string;
  label: string;
  securityLabel: string;
  securityFg: string;
  securityBg: string;
  canRead: boolean;
  canWrite: boolean;
  derived: boolean;
  moderation: boolean;
  moderationLabel: string;
  nodes: Node[];
};

//var treeData = {"id":"?"} as Node;

// This is where we store the tree, as we load it
var treeData = {
  "id": "/files/",
  "label": "files",
  "securityLabel": "PUBLIC",
  "securityFg": "white",
  "securityBg": "green",
  "canRead": true,
  "canWrite": false,
  "derived": false,
  "moderation": false,
  "moderationLabel": "",
  "nodes": []
} as Node;

var endpoint = "http://localhost:9321";


function labeledNode(node: Node) {
  return (
    <>
    <span style={{
      backgroundColor: node.securityBg, 
      color: node.securityFg, 
      opacity: 100,
    }}>
      {node.securityLabel}&nbsp;
      {node.canRead ? 'R' : ''}
      {node.canWrite ? 'W' : ''}
      {node.moderation ? '!!' : ''}
    </span>
    &nbsp;
    <span>{node.label}</span>
    </>
  );
}

function renderTree(nodes : Node) {
  return (
    <TreeItem key={nodes.id} nodeId={nodes.id} label={labeledNode(nodes)}>
      {Array.isArray(nodes.nodes) ? nodes.nodes.map((node) => renderTree(node)) : null}
    </TreeItem>
  );
}

// Maybe make our json match Material UI's TreeView
function assignNode(v: string) : Node {
  var parsedv = JSON.parse(v);
  var td = {} as Node;
  td["id"] = parsedv["path"] + parsedv["name"];
  td["label"] = parsedv["name"];
  if(parsedv["isDir"]) {
    td["id"] += "/";
  }
  td["securityLabel"] = parsedv["attributes"]["Label"];
  td["securityFg"] = parsedv["attributes"]["LabelFg"];
  td["securityBg"] = parsedv["attributes"]["LabelBg"];
  td["canRead"] = parsedv["attributes"]["Read"];
  td["canWrite"] = parsedv["attributes"]["Write"];
  td["derived"] = parsedv["attributes"]["Derived"];
  td["moderation"] = parsedv["attributes"]["Moderation"];
  td["moderationLabel"] = parsedv["attributes"]["ModerationLabel"];
  td["nodes"] = [];
  return td;
}

async function getTreeNode(fsPath: string,setTreeNode:(n:Node)=>void) {
  try {
    const response = await fetch(endpoint + fsPath );
    const data = await response.text();
    var n = assignNode(data);
    setTreeNode(n);
    console.log("Set node "+JSON.stringify(n));
  } catch(err) {
    console.log(err);
  }
}


function FullTreeView() : JSX.Element {
  //TODO: figure out responding to expand events
  getTreeNode("/files/?json=true",(n:Node) => { 
    treeData = n;
    console.log("Set node "+n.id); 
  });
  return (
    <TreeView
      aria-label="file system navigator"
      defaultCollapseIcon={<ExpandMoreIcon />}
      defaultExpandIcon={<ChevronRightIcon />}
      style={{ alignContent: 'left', textAlign: 'left', height: 240, flexGrow: 1, maxWidth: 400, overflowY: 'auto' }}   
    >
      {renderTree(treeData)}
    </TreeView>
  );
}


function App() {
  return (
    <div className="App">
      <header className="App-header">
        <h1>MicroCMS</h1>
        <FullTreeView/>
      </header>
    </div>
  );
}

export default App;
