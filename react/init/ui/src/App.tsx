import React from 'react';
//import logo from './logo.svg';
import './App.css';
//import launch from './launch.svg';
import logo from './logo.svg';

import { useState } from 'react';
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
  Moderation?: boolean;
  ModerationLabel?: string;
}

interface SNode {
  name: string;
  path: string;
  isDir: boolean;
  size?: number;
  //context?: string;
  part?: number;
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
  size?: number;
  part?: number;
  context?: string;
  derived?: boolean;
  moderation?: boolean;
  moderationLabel?: string;
  children: string[];
};

type Nodes = {
  [id: string]: Node;
};


interface Hit {
  id: string;
  children: string[];
}

type Hits = {
  [id: string]: Hit;
}

type HideableNodes = {
  nodes : Nodes;
  hideMismatches : boolean;
  hideDerived : boolean;
};

type Endpoints = {
  prefixes: { 
    [id: string]: string 
  }
};

var endpoints : Endpoints;


function getEndpoint(name: string) : string {
  return endpoints["prefixes"][name];
}

function doesMatchQuery(node: Node, query: Hits) : boolean {
  // cant match empty
  if(Object.keys(query).length === 0) {
    return false;
  }
  // exact match is easy
  if(query[node.id]) {
    return true;
  }
  var answer = false;
  // parent match if our key is a substring of one in query
  Object.keys(query).forEach(function(k) {
    if(k.startsWith(node.id)) {
      answer = true;
    }
  });
  return answer;
}

function convertHit(p: SNode) : Hit {
  var td = {} as Hit;
  td.id = p.path + p.name;
  td.children = [];
  return td;
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
  td.isDir = p.isDir;
  var a = p.attributes;
  td.securityLabel = a.Label;
  td.securityFg = a.LabelFg;
  td.securityBg = a.LabelBg;
  td.canRead = a.Read;
  td.canWrite = a.Write;
  td.part = p.part;
  td.size = p.size;
  td.matchesQuery = false;
  //td.context = p.context;
  // XXX hack
  td.derived = ((td.label.indexOf("--")>0) ? true : false);
  td.moderation = a.Moderation ? true : false;
  td.moderationLabel = a.ModerationLabel ? a.ModerationLabel : "";
  td.matchesQuery = false;
  td.children = [];
  return td;
}

function matchTreeState(nodes: Nodes, query: Hits) {
  Object.keys(nodes).forEach(function(k) {
    nodes[k].matchesQuery = doesMatchQuery(nodes[k],query);
  });
}

// Update the tree state
function convertTreeState(p: SNode, nodes: Nodes):Nodes {
  var n = convertNode(p);
  nodes[n.id] = n;
  if(p.isDir && p.children) {
    for(var i=0; i<p.children.length; i++) {
      var c = convertNode(p.children[i])
      nodes[c.id] = c;
      nodes[n.id].children.push(c.id);
    }
  }
  return nodes;
}

function convertSearchState(p: SNode, nodes: Hits):Hits {
  var n = convertHit(p);
  nodes[n.id] = n;
  if(p.isDir && p.children) {
    for(var i=0; i<p.children.length; i++) {
      var c = convertHit(p.children[i]);
      nodes[c.id] = c;
      nodes[n.id].children.push(c.id);
    }
  }
  return nodes;
}

function asSize(size: number) : string {
  var s = size;
  var units = ["B","KB","MB","GB","TB","PB","EB","ZB","YB"];
  var i = 0;
  while(s>1024) {
    s = s/1024;
    i++;
  }
  return s.toFixed(2) + " " + units[i];
}

function LabeledNode(nodes: Nodes, node: Node) : JSX.Element {
    var thumbnail = node.id+"--thumbnail.png";
    var color="white";
    var textNotMatched = "#a0a0a0";
    if(!node.matchesQuery) {
      color = textNotMatched;
    }
    var note = "";
    if(node.moderation && !node.derived) {
      color = "#ff4040";
      if(!node.matchesQuery) {
        color = "#a03030";
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
    
    var theImg = <></>;
    if(nodes[node.id+"--thumbnail.png"]) {
      theImg = <img 
        src={thumbnail} 
        height="26"
        width="auto" 
        alt="" 
        style={{verticalAlign:'bottom', border: '0px', objectFit: 'cover'}}
        onMouseOver={e => (e.currentTarget.height=200)}
        onMouseOut={e => (e.currentTarget.height=20)}
      />;
    }

    var nodeSize = asSize(node.size ? node.size : 0);

    var theText = 
    <a 
      href={node.id} 
      target="_blank"
      rel="noreferrer" 
      style={{color:color, textDecoration:'none'}}
    >
      {node.label}&nbsp;
      {note}
      &nbsp;
      ({nodeSize})
      &nbsp;
      {theImg}
    </a>

    if(node.isDir) {
      thumbnail = "";
      color="white";
      if(!node.matchesQuery) {
        color = textNotMatched;
      }
      var launchIcon = logo;
      if(!nodes[node.id + "index.html"]) {
        launchIcon = "";
      }
      theText = 
      <span style={{color:color, textDecoration:'none'}}>
        {node.label} 
        &nbsp; 
        <a href={node.id} target="_blank" rel="noreferrer">
          <img height="20" alt="tbd" width="auto" style={{verticalAlign: 'bottom'}} src={launchIcon}/>
        </a>
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
      {node.moderation ? ' !!' : ''}
    </span>
    &nbsp;
    <span>{theText}</span>
    </div>
  );
};


function SearchableTreeView() : JSX.Element {
  const [hideMismatchedData, setHideMismatchedData] = useState<boolean>(false);
  const [hideDerivedData, setHideDerivedData] = useState<boolean>(true);
  const [searchData, setSearchData] = useState<Hits>({});
  const [hideableData, setHideableData] = useState<HideableNodes>({
    hideMismatches: false,
    hideDerived: true,
    nodes: {
      "/files/": {
        id:"/files/",
        label:"files/",
        isDir:true,
        securityLabel:"HOME",
        securityFg:"white",
        securityBg:"darkblue",
        matchesQuery: false, 
        canRead:true,
        canWrite:false,
        children:[]
      }
    }
}); 

const detectKeys = async (e : React.KeyboardEvent<HTMLInputElement>) => {
  try {
    if(e.key === "Enter") {
      // Get the keyword and the root to search at
      var searchTerms = e.currentTarget.value;
      var keywords = searchTerms
      var searchAt = "/files/";
      if(searchTerms.startsWith("/files/") ) {
        var tokens = searchTerms.split(" ");
        searchAt = tokens[0];
        keywords = tokens.slice(1).join(" ");
      }
      // Clean the tree before the query
      const response = await fetch(
        getEndpoint("microcms") + "/search"+searchAt+"?json=true&hideContent&match="+keywords,
        {"credentials": "same-origin"},
      );
      const p = await response.json() as SNode;
      var existingSearchData = {} as Hits;
      var newSearchData = convertSearchState(p, existingSearchData);
      setSearchData({...newSearchData});
      matchTreeState(hideableData.nodes,newSearchData);
      setHideableData({...{
        nodes: hideableData.nodes, 
        hideMismatches: hideMismatchedData,
        hideDerived: hideDerivedData
      }});
    }
  } catch (error) {
    console.error('Error fetching data:', error);
  }
  };
  
  const loadTreeItem = async (node: Node) => {
    try {
      // note: we can delete collapsed nodes to save memory
      if(node.isDir && node.id.endsWith("/")) {
        const response = await fetch(
          getEndpoint("microcms") + node.id + "?json=true&listing=true",
          {"credentials": "same-origin"},
        );
        const p = await response.json() as SNode;
        var newTreeData = convertTreeState(p, hideableData.nodes);
        matchTreeState(newTreeData,searchData);
        setHideableData({...{
          nodes: newTreeData, 
          hideMismatches: hideMismatchedData,
          hideDerived: hideDerivedData
        }});
      }
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  const handleIconClick = async (e: React.MouseEvent<Element,MouseEvent>,node: Node) => {
    loadTreeItem(node);
  };

  const handleLabelClick = async (e: React.MouseEvent<Element,MouseEvent>,node: Node) => {
    loadTreeItem(node);
  };
  
  const clickHideMismatch = async (e: React.SyntheticEvent<Element,Event>) => {
    try {
      setHideMismatchedData(!hideMismatchedData);
      setHideableData({...{
        nodes: hideableData.nodes, 
        hideMismatches: !hideMismatchedData,
        hideDerived: hideDerivedData
      }});
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  const clickHideDerived = async (e: React.SyntheticEvent<Element,Event>) => {
    try {
      setHideDerivedData(!hideDerivedData);
      setHideableData({...{
        nodes: hideableData.nodes, 
        hideMismatches: hideMismatchedData,
        hideDerived: !hideDerivedData
      }});
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };

  var renderTree = function(nodes : Nodes, id:string) : JSX.Element {
    var matches = nodes[id].matchesQuery? true : false;;
    var hidden = (!matches && hideMismatchedData) || 
      (nodes[id].derived && hideDerivedData) ||
      (id.endsWith("/permissions.rego") && hideDerivedData)
    ;
    return (
      <TreeItem 
        key={id}
        nodeId={id}
        hidden={hidden} 
        label={LabeledNode(nodes, nodes[id])}
        onIconClick={e => handleIconClick(e,nodes[id])}
        onLabelClick={e => handleLabelClick(e,nodes[id])}
      >
        {Array.isArray(nodes[id].children) ? nodes[id].children.map((v) => renderTree(nodes,v)) : null}
      </TreeItem>
    );
  };

  return (
    <>
    <div style={{padding: 10}}>
      <span 
        style={{fontSize: 24}}
      >
          Search:
      </span>
        &nbsp; 
        <input 
          type="text" 
          name="match" 
          size={60} 
          onKeyUp={e => detectKeys(e)} 
        />
      <br/>
      <i style={{fontSize: 14, color: 'gray'}}>search like: /files/rob cat OR dog. Press 'Enter' or 'Return' to run query.</i>
      <div style={{fontSize: 14, color: 'gray'}}>
        &nbsp; 
        <input 
          type="checkbox" 
          name="hidemismatch" 
          value="/files/" 
          checked={hideMismatchedData}
          onClick={e => clickHideMismatch(e)}     
        /> Hide Mismatch &nbsp; 
      </div>  
      <div style={{fontSize: 14, color: 'gray'}}>
        &nbsp; 
        <input 
          type="checkbox" 
          name="hidederived" 
          value="/files/" 
          checked={hideDerivedData}
          onClick={e => clickHideDerived(e)}     
        /> Hide Metadata &nbsp;
      </div>  
    </div>
    <TreeView      
      aria-label="file system navigator"
      defaultCollapseIcon={<ExpandMoreIcon />}
      defaultExpandIcon={<ChevronRightIcon />}
    >
      {renderTree(hideableData.nodes,"/files/")}
    </TreeView>
    </>
  );
}


function App() {

  const loadEndpoints = async () => {
    const response = await fetch(
      "./endpoints.json",
      {"credentials": "same-origin"},
    );      
    endpoints = await response.json() as Endpoints;
  };
  loadEndpoints();

  return (
    <div 
      style={{ 
        color: 'white', 
        background: 'black', 
        alignContent: 'left', 
        textAlign: 'left', 
        width: "100%", 
        height: 'auto', 
        flexGrow: 1, 
        overflow: 'flex'
      }}   
    >
      <SearchableTreeView/>
    </div>
  );
}

export default App;
