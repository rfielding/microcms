import React from 'react';
//import logo from './logo.svg';
import './App.css';

import { useState, useEffect } from 'react';
import axios from 'axios';

import TreeView from '@material-ui/lab/TreeView';
import TreeItem from '@material-ui/lab/TreeItem';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';

interface Attributes {
  Label: string;
  LabelFg: string;
  LabelBg: string;
  Read: boolean;
  Write: boolean;
  Derived?: boolean;
  Moderation?: boolean;
  ModerationLabel?: string;
}

interface SNode {
  name: string;
  path: string;
  isDir: boolean;
  size?: number;
  attributes: Attributes;
  children?: SNode[];
};

interface Node {
  id: string;
  label: string;
  securityLabel: string;
  securityFg: string;
  securityBg: string;
  canRead: boolean;
  canWrite: boolean;
  derived?: boolean;
  moderation?: boolean;
  moderationLabel?: string;
  children: string[];
};

type Nodes = {
  [id: string]: Node;
};

// This is where we store the tree, as we load it
var treeData = {
  "/files/": {
    id:"/files/",
    label:"files/",
    securityLabel:"PUBLIC",
    securityFg:"white",
    securityBg:"green",
    canRead:true,
    canWrite:false,
    derived:false,
    moderation:false,
    moderationLabel:"",
    children:[
      "/files/init/",
      "/files/permissions.rego"
    ]
  },
  "/files/init/":{
    id:"/files/init/",
    label:"init/",
    securityLabel:"PUBLIC",
    securityFg:"white",
    securityBg:"green",
    canRead:true,
    canWrite:false,
    derived:false,
    moderation:false,
    moderationLabel:"",
    children:[]
  },
  "/files/permissions.rego": {
    id:"/files/permissions.rego",
    label:"permissions.rego",
    securityLabel:"PUBLIC",
    securityFg:"white",
    securityBg:"green",
    canRead:true,
    canWrite:false,
    derived:false,
    moderation:false,
    moderationLabel:"",
    children:[]
  }
} as Nodes;

// XXX: along with busting open CORS ... 
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

function renderTree(node : Node) {
  return (
    <TreeItem key={node.id} nodeId={node.id} label={labeledNode(node)}>
      {Array.isArray(node.children) ? node.children.map((id) => renderTree(treeData[id])) : null}
    </TreeItem>
  );
}

// Maybe make our json match Material UI's TreeView
function convertNode(p: SNode) : Node {
  var td = {} as Node;
  td.id = p.path + p.name;
  td.label = p.name;
  if(p.isDir) {
    td.id += "/";
    td.label += "/";
  }
  var a = p.attributes;
  if(a === undefined) {
    console.log("No attributes for "+JSON.stringify(p));
  }
  td.securityLabel = a.Label;
  td.securityFg = a.LabelFg;
  td.securityBg = a.LabelBg;
  td.canRead = a.Read;
  td.canWrite = a.Write;
  td.derived = a.Derived ? true : false;
  td.moderation = a.Moderation ? true : false;
  td.moderationLabel = a.ModerationLabel ? a.ModerationLabel : "";
  td.children = [];
  return td;
}

// Update the tree state
function updateTreeState(v: string) {
  var p = JSON.parse(v) as SNode;
  var n = convertNode(p);
  if(p.isDir && p.children) {
    for(var i=0; i<p.children.length; i++) {
      var c = convertNode(p.children[i])
      treeData[c.id] = c;
      treeData[n.id].children.push(c.id);
    }
  }
}

async function fetchNode(fullName: string) {
  var url = endpoint + fullName + "?json=true";
  try {
    console.log("Fetching "+fullName);
    if(fullName.endsWith("/") && fullName != "/") {
      const response = await fetch(
        url,
        {credentials: "same-origin"}
      );
      const data = await response.text();
      updateTreeState(data);
      //console.log("Set node "+JSON.stringify(treeData));
    }
  } catch(err) {
    console.log("while fetching "+url+" "+err);
  }
}


function FullTreeView() : JSX.Element {
  const selected = (event: React.ChangeEvent<{}>, value: string[]) => {
    const p = value+"";
    console.log("Selected "+p);
    fetchNode(p);
  };
  return (
    <TreeView      
      onNodeSelect={(event: React.ChangeEvent<{}>, value: string[]) => selected(event, value)}
      aria-label="file system navigator"
      defaultCollapseIcon={<ExpandMoreIcon />}
      defaultExpandIcon={<ChevronRightIcon />}
      style={{ alignContent: 'left', textAlign: 'left', height: 240, flexGrow: 1, maxWidth: 400, overflowY: 'auto' }}   
    >
      {renderTree(treeData["/files/"])}
    </TreeView>
  );
}


function App() {
  return (
    <div className="App">
      <header className="App-header" style={{alignContent:'left'}}>
        <FullTreeView/>
      </header>
    </div>
  );
}

export default App;
