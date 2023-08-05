import React from 'react';
//import logo from './logo.svg';
import './App.css';

import { useState, useEffect } from 'react';

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
  isDir: boolean;
  securityLabel: string;
  securityFg: string;
  securityBg: string;
  canRead: boolean;
  canWrite: boolean;
  matchesQuery: boolean;
  derived?: boolean;
  moderation?: boolean;
  moderationLabel?: string;
  children: string[];
};

type Nodes = {
  [id: string]: Node;
};

type NodesStart = {
  nodes: Nodes;
  start: string;
};



// XXX: along with busting open CORS ... 
var endpoint = "http://localhost:9321";






// Maybe make our json match Material UI's TreeView
function convertNode(p: SNode) : Node {
  var td = {} as Node;
  td.id = p.path + p.name;
  td.label = p.name;
  if(p.isDir) {
    td.id += "/";
    td.label += "/";
  }
  td.isDir = p.isDir;
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
function convertTreeState(p: SNode, ins: NodesStart):NodesStart {
  var n = convertNode(p);
  ins.nodes[n.id] = n;
  if(p.isDir && p.children) {
    for(var i=0; i<p.children.length; i++) {
      var c = convertNode(p.children[i])
      ins.nodes[c.id] = c;
      ins.nodes[n.id].children.push(c.id);
    }
  }
  return ins;
}

function LabeledNode(node: Node) : JSX.Element {
    var thumbnail = node.id+"--thumbnail.png";
    var color="white";
    if(!node.matchesQuery) {
      color = "gray";
    }
    var note = "";
    if(node.moderation && !node.derived) {
      color = "red";
      if(!node.matchesQuery) {
        color = "darkred";
      }
      note = " ( "+node.moderationLabel+" )";
    }
    if(node.derived) {
      thumbnail = "";
      color="gray";
      if(!node.matchesQuery) {
        color = "darkgray";
      }
    }

    var theText = 
    <a href={node.id} target="_blank" style={{color:color, textDecoration:'none'}}>
      {node.label}&nbsp;
      <img src={thumbnail} height="20" width="auto" alt="" style={{verticalAlign:'center'}}/>
      {note}
    </a>

    if(node.isDir) {
      thumbnail = "";
      color="white";
      if(!node.matchesQuery) {
        color = "gray";
      }
      theText = 
      <span style={{color:color, textDecoration:'none'}}>
        {node.label}
      </span>;

    }
  
  return (
    <div>
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
    <span>{theText}</span>
    </div>
  );
};


function FullTreeView() : JSX.Element {
  const [treeData, setTreeData] = useState<NodesStart>({
    start: "/files/",
    nodes: {
      "/files/": {
        id:"/files/",
        label:"files/",
        isDir:true,
        securityLabel:"PUBLIC",
        securityFg:"white",
        securityBg:"green",
        matchesQuery: false,
        canRead:true,
        canWrite:false,
        children:[]
      }
    }
  });
  
  const handleToggle = async (node: Node) => {
    try {
      if(node.id.endsWith("/")) {
        const response = await fetch(
          endpoint + node.id + "?json=true&listing=true",
          {"credentials": "same-origin"},
        );
        const p = await response.json() as SNode;
        var newTreeData = convertTreeState(p, treeData);
        setTreeData({...newTreeData});
      }
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  const handleClick = async (node: Node) => {
    if(node.isDir) {
      await handleToggle(node);
    } else {
      console.log("Clicked on "+node.id);
    }
  };
  
  var renderTree = function(ins : NodesStart, id:string) : JSX.Element {
    return (
      <TreeItem 
        nodeId={id} 
        label={LabeledNode(ins.nodes[id])}
        onIconClick={() => handleToggle(ins.nodes[id])}
        onLabelClick={() => handleClick(ins.nodes[id])}
      >
        {Array.isArray(ins.nodes[id].children) ? ins.nodes[id].children.map((v) => renderTree(ins,v)) : null}
      </TreeItem>
    );
  };

  return (
    <TreeView      
      aria-label="file system navigator"
      defaultCollapseIcon={<ExpandMoreIcon />}
      defaultExpandIcon={<ChevronRightIcon />}
    >
      {renderTree(treeData,"/files/")}
    </TreeView>
  );
}


function App() {
  return (
    <div 
      className="App" 
      style={{ 
        color: 'white', 
        background: 'black', 
        alignContent: 'left', 
        textAlign: 'left', 
        width: 1040, 
        height: 1000, 
        flexGrow: 0, 
        overflow: 'auto' 
      }}   
    >
      <FullTreeView/>
    </div>
  );
}

export default App;
