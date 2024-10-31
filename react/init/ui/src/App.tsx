import React from 'react';
import './App.css';
import logo from './logo.svg';

import { useState, DragEvent, ChangeEvent, useEffect } from 'react';
import TreeView from '@material-ui/lab/TreeView';
import TreeItem from '@material-ui/lab/TreeItem';
import ExpandMoreIcon from '@material-ui/icons/ExpandMore';
import ChevronRightIcon from '@material-ui/icons/ChevronRight';
import { Upload, X } from 'lucide-react';

interface FileUploadProps {
  onUploadComplete?: (file: File, targetUrl: string) => void;
  onDeleteComplete?: (targetUrl: string) => void;
  onError?: (error: Error) => void;
}

const FileUpload: React.FC<FileUploadProps> = ({
  onUploadComplete,
  onDeleteComplete,
  onError,
}) => {
  const [isDragging, setIsDragging] = useState<boolean>(false);
  const [file, setFile] = useState<File | null>(null);
  const [targetUrl, setTargetUrl] = useState<string>('');
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState<boolean>(false);
  const [isDeleting, setIsDeleting] = useState<boolean>(false);

  const handleDragOver = (e: DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>): void => {
    e.preventDefault();
    setIsDragging(false);
    
    const droppedFile = e.dataTransfer.files[0];
    if (droppedFile) {
      handleFile(droppedFile);
    }
  };

  const handleFileInput = (e: ChangeEvent<HTMLInputElement>): void => {
    if (!e.target.files) return;
    const selectedFile = e.target.files[0];
    if (selectedFile) {
      handleFile(selectedFile);
    }
  };

  const handleFile = (newFile: File): void => {
    setFile(newFile);
    // Get the current path and append the new filename
    const currentPath = (document.getElementById('targetUrl') as HTMLInputElement)?.value || '';
    setTargetUrl(currentPath + newFile.name);
    setUploadProgress(0);
  };

  const handleTargetUrlChange = (e: ChangeEvent<HTMLInputElement>): void => {
    setTargetUrl(e.target.value);
  };

  const handleUpload = async (): Promise<void> => {
    if (!file || !targetUrl) return;

    try {
      setIsUploading(true);
      const formData = new FormData();
      formData.append('file', file);

      const response = await fetch(targetUrl, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error(`Upload failed: ${response.statusText}`);
      }

      onUploadComplete?.(file, targetUrl);
    } catch (error) {
      onError?.(error as Error);
    } finally {
      setIsUploading(false);
    }
  };

  const handleDelete = async (): Promise<void> => {
    if (!targetUrl) return;

    try {
      setIsDeleting(true);
      const response = await fetch(targetUrl, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error(`Delete failed: ${response.statusText}`);
      }

      onDeleteComplete?.(targetUrl);
      // Optionally clear the form after successful delete
      // setFile(null);
      // setTargetUrl('');
    } catch (error) {
      onError?.(error as Error);
    } finally {
      setIsDeleting(false);
    }
  };

  const removeFile = (): void => {
    setFile(null);
    setUploadProgress(0);
  };

  const isUploadReady = file && targetUrl.trim();

  return (
    <div className="w-full max-w-md mx-auto p-6">
      <div className="mb-4">
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Target URL
        </label>
        <div className="flex gap-2">
          <input
            type="text"
            id="targetUrl"
            value={targetUrl}
            onChange={handleTargetUrlChange}
            className="min-w[800px] flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            placeholder="Enter target URL"
            aria-label="Target URL"
          ></input>

          {targetUrl && (
            <button
              onClick={handleDelete}
              disabled={isDeleting || isUploading}
              className={`px-4 py-2 text-white rounded-lg ${
                isDeleting
                  ? 'bg-red-400 cursor-not-allowed'
                  : 'bg-red-500 hover:bg-red-600'
              }`}
              aria-label="Delete at target URL"
            >
              {isDeleting ? 'Deleting...' : 'Delete'}
            </button>
          )}
        </div>
      </div>
      {isUploadReady && (
        <button
          onClick={handleUpload}
          disabled={isUploading || isDeleting}
          className={`mt-4 w-full px-4 py-2 text-white rounded-lg ${
            isUploading
              ? 'bg-blue-400 cursor-not-allowed'
              : 'bg-blue-500 hover:bg-blue-600'
          }`}
        >
          {isUploading ? 'Uploading...' : 'Upload File'}
        </button>
      )}

      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center ${
          isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300'
        }`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        role="presentation"
      >
        {!file ? (
          <>
            <Upload className="mx-auto h-12 w-12 text-gray-400" />
            <p className="mt-2 text-gray-600">
              Drag and drop a file here, or
            </p>
            <label className="mt-4 inline-block">
              <span className="px-4 py-2 bg-blue-500 text-white rounded-lg cursor-pointer hover:bg-blue-600">
                Browse Files
              </span>
              <input
                type="file"
                className="hidden"
                onChange={handleFileInput}
                aria-label="File upload"
              />
            </label>
          </>
        ) : (
          <div className="space-y-4">
            <div className="bg-gray-50 p-4 rounded-lg flex items-center">
              <div className="flex-1">
                <p className="text-sm font-medium text-gray-900">{file.name}</p>
                {isUploading && (
                  <div className="mt-1 h-2 w-full bg-gray-200 rounded-full">
                    <div
                      className="h-2 bg-blue-500 rounded-full transition-all duration-500"
                      style={{ width: `${uploadProgress}%` }}
                      role="progressbar"
                      aria-valuenow={uploadProgress}
                      aria-valuemin={0}
                      aria-valuemax={100}
                    />
                  </div>
                )}
              </div>
              <button
                onClick={removeFile}
                className="ml-4 text-gray-400 hover:text-gray-500"
                aria-label="Remove file"
                disabled={isUploading}
              >
                <X className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}
      </div>


      {file && !targetUrl && (
        <p className="mt-2 text-sm text-red-500">
          Please enter a target URL to enable upload
        </p>
      )}
    </div>
  );
};

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

type Me = {
  id: string[];
  name: string[];
  email: string[];
  roles: string[];
  permissions: string[];
};

var me : Me;

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
          <img height="20" alt="" width="auto" style={{verticalAlign: 'bottom'}} src={launchIcon}/>
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
    if (node.isDir) {
      const targetUrlInput = document.getElementById('targetUrl') as HTMLInputElement;
      if (targetUrlInput) {
        targetUrlInput.value = node.id;
        // Trigger change event to update React state
        targetUrlInput.dispatchEvent(new Event('change', { bubbles: true }));
      }
    }    
  };

  const handleLabelClick = async (e: React.MouseEvent<Element,MouseEvent>,node: Node) => {
    handleIconClick(e,node);
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

  const loadMe = async () => {
    const response = await fetch(
      "./me",
      {"credentials": "same-origin"},
    );
    me = await response.json() as Me;
    var who = "anonymous";
    if(me.email && me.email.length>0) {
      who = me.email[0];
    }
    document.getElementById('mespan')!.innerHTML = who;
  };
  loadMe();

  return (
    <>
    <a href="/me" style={{fontSize: 14, color: 'gray', textDecoration: 'none'}}>As </a>
    <span id='mespan'></span>

    <div style={{padding: 10}}>
      <span 
        style={{fontSize: 16}}
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
      <hr/>
      <FileUpload/>
    </div>
  );
}

export default App;
