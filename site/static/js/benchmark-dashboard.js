var lO=Object.defineProperty;var Yg=e=>{throw TypeError(e)};var uO=(e,t,n)=>t in e?lO(e,t,{enumerable:!0,configurable:!0,writable:!0,value:n}):e[t]=n;var Qg=(e,t,n)=>uO(e,typeof t!="symbol"?t+"":t,n),$d=(e,t,n)=>t.has(e)||Yg("Cannot "+n);var z=(e,t,n)=>($d(e,t,"read from private field"),n?n.call(e):t.get(e)),me=(e,t,n)=>t.has(e)?Yg("Cannot add the same private member more than once"):t instanceof WeakSet?t.add(e):t.set(e,n),ne=(e,t,n,r)=>($d(e,t,"write to private field"),r?r.call(e,n):t.set(e,n),n),gt=(e,t,n)=>($d(e,t,"access private method"),n);var $u=(e,t,n,r)=>({set _(a){ne(e,t,a,n)},get _(){return z(e,t,r)}});function cO(e,t){for(var n=0;n<t.length;n++){const r=t[n];if(typeof r!="string"&&!Array.isArray(r)){for(const a in r)if(a!=="default"&&!(a in e)){const o=Object.getOwnPropertyDescriptor(r,a);o&&Object.defineProperty(e,a,o.get?o:{enumerable:!0,get:()=>r[a]})}}}return Object.freeze(Object.defineProperty(e,Symbol.toStringTag,{value:"Module"}))}var Mu=typeof globalThis<"u"?globalThis:typeof window<"u"?window:typeof global<"u"?global:typeof self<"u"?self:{};function _e(e){return e&&e.__esModule&&Object.prototype.hasOwnProperty.call(e,"default")?e.default:e}var z2={exports:{}},yf={},H2={exports:{}},ce={};/**
 * @license React
 * react.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var yu=Symbol.for("react.element"),pO=Symbol.for("react.portal"),fO=Symbol.for("react.fragment"),dO=Symbol.for("react.strict_mode"),mO=Symbol.for("react.profiler"),hO=Symbol.for("react.provider"),vO=Symbol.for("react.context"),yO=Symbol.for("react.forward_ref"),gO=Symbol.for("react.suspense"),xO=Symbol.for("react.memo"),wO=Symbol.for("react.lazy"),Zg=Symbol.iterator;function bO(e){return e===null||typeof e!="object"?null:(e=Zg&&e[Zg]||e["@@iterator"],typeof e=="function"?e:null)}var G2={isMounted:function(){return!1},enqueueForceUpdate:function(){},enqueueReplaceState:function(){},enqueueSetState:function(){}},U2=Object.assign,W2={};function ns(e,t,n){this.props=e,this.context=t,this.refs=W2,this.updater=n||G2}ns.prototype.isReactComponent={};ns.prototype.setState=function(e,t){if(typeof e!="object"&&typeof e!="function"&&e!=null)throw Error("setState(...): takes an object of state variables to update or a function which returns an object of state variables.");this.updater.enqueueSetState(this,e,t,"setState")};ns.prototype.forceUpdate=function(e){this.updater.enqueueForceUpdate(this,e,"forceUpdate")};function q2(){}q2.prototype=ns.prototype;function kv(e,t,n){this.props=e,this.context=t,this.refs=W2,this.updater=n||G2}var Cv=kv.prototype=new q2;Cv.constructor=kv;U2(Cv,ns.prototype);Cv.isPureReactComponent=!0;var Jg=Array.isArray,V2=Object.prototype.hasOwnProperty,_v={current:null},K2={key:!0,ref:!0,__self:!0,__source:!0};function X2(e,t,n){var r,a={},o=null,i=null;if(t!=null)for(r in t.ref!==void 0&&(i=t.ref),t.key!==void 0&&(o=""+t.key),t)V2.call(t,r)&&!K2.hasOwnProperty(r)&&(a[r]=t[r]);var s=arguments.length-2;if(s===1)a.children=n;else if(1<s){for(var l=Array(s),u=0;u<s;u++)l[u]=arguments[u+2];a.children=l}if(e&&e.defaultProps)for(r in s=e.defaultProps,s)a[r]===void 0&&(a[r]=s[r]);return{$$typeof:yu,type:e,key:o,ref:i,props:a,_owner:_v.current}}function PO(e,t){return{$$typeof:yu,type:e.type,key:t,ref:e.ref,props:e.props,_owner:e._owner}}function Av(e){return typeof e=="object"&&e!==null&&e.$$typeof===yu}function SO(e){var t={"=":"=0",":":"=2"};return"$"+e.replace(/[=:]/g,function(n){return t[n]})}var e1=/\/+/g;function Md(e,t){return typeof e=="object"&&e!==null&&e.key!=null?SO(""+e.key):t.toString(36)}function xc(e,t,n,r,a){var o=typeof e;(o==="undefined"||o==="boolean")&&(e=null);var i=!1;if(e===null)i=!0;else switch(o){case"string":case"number":i=!0;break;case"object":switch(e.$$typeof){case yu:case pO:i=!0}}if(i)return i=e,a=a(i),e=r===""?"."+Md(i,0):r,Jg(a)?(n="",e!=null&&(n=e.replace(e1,"$&/")+"/"),xc(a,t,n,"",function(u){return u})):a!=null&&(Av(a)&&(a=PO(a,n+(!a.key||i&&i.key===a.key?"":(""+a.key).replace(e1,"$&/")+"/")+e)),t.push(a)),1;if(i=0,r=r===""?".":r+":",Jg(e))for(var s=0;s<e.length;s++){o=e[s];var l=r+Md(o,s);i+=xc(o,t,n,l,a)}else if(l=bO(e),typeof l=="function")for(e=l.call(e),s=0;!(o=e.next()).done;)o=o.value,l=r+Md(o,s++),i+=xc(o,t,n,l,a);else if(o==="object")throw t=String(e),Error("Objects are not valid as a React child (found: "+(t==="[object Object]"?"object with keys {"+Object.keys(e).join(", ")+"}":t)+"). If you meant to render a collection of children, use an array instead.");return i}function Ru(e,t,n){if(e==null)return e;var r=[],a=0;return xc(e,r,"","",function(o){return t.call(n,o,a++)}),r}function OO(e){if(e._status===-1){var t=e._result;t=t(),t.then(function(n){(e._status===0||e._status===-1)&&(e._status=1,e._result=n)},function(n){(e._status===0||e._status===-1)&&(e._status=2,e._result=n)}),e._status===-1&&(e._status=0,e._result=t)}if(e._status===1)return e._result.default;throw e._result}var Dt={current:null},wc={transition:null},kO={ReactCurrentDispatcher:Dt,ReactCurrentBatchConfig:wc,ReactCurrentOwner:_v};function Y2(){throw Error("act(...) is not supported in production builds of React.")}ce.Children={map:Ru,forEach:function(e,t,n){Ru(e,function(){t.apply(this,arguments)},n)},count:function(e){var t=0;return Ru(e,function(){t++}),t},toArray:function(e){return Ru(e,function(t){return t})||[]},only:function(e){if(!Av(e))throw Error("React.Children.only expected to receive a single React element child.");return e}};ce.Component=ns;ce.Fragment=fO;ce.Profiler=mO;ce.PureComponent=kv;ce.StrictMode=dO;ce.Suspense=gO;ce.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED=kO;ce.act=Y2;ce.cloneElement=function(e,t,n){if(e==null)throw Error("React.cloneElement(...): The argument must be a React element, but you passed "+e+".");var r=U2({},e.props),a=e.key,o=e.ref,i=e._owner;if(t!=null){if(t.ref!==void 0&&(o=t.ref,i=_v.current),t.key!==void 0&&(a=""+t.key),e.type&&e.type.defaultProps)var s=e.type.defaultProps;for(l in t)V2.call(t,l)&&!K2.hasOwnProperty(l)&&(r[l]=t[l]===void 0&&s!==void 0?s[l]:t[l])}var l=arguments.length-2;if(l===1)r.children=n;else if(1<l){s=Array(l);for(var u=0;u<l;u++)s[u]=arguments[u+2];r.children=s}return{$$typeof:yu,type:e.type,key:a,ref:o,props:r,_owner:i}};ce.createContext=function(e){return e={$$typeof:vO,_currentValue:e,_currentValue2:e,_threadCount:0,Provider:null,Consumer:null,_defaultValue:null,_globalName:null},e.Provider={$$typeof:hO,_context:e},e.Consumer=e};ce.createElement=X2;ce.createFactory=function(e){var t=X2.bind(null,e);return t.type=e,t};ce.createRef=function(){return{current:null}};ce.forwardRef=function(e){return{$$typeof:yO,render:e}};ce.isValidElement=Av;ce.lazy=function(e){return{$$typeof:wO,_payload:{_status:-1,_result:e},_init:OO}};ce.memo=function(e,t){return{$$typeof:xO,type:e,compare:t===void 0?null:t}};ce.startTransition=function(e){var t=wc.transition;wc.transition={};try{e()}finally{wc.transition=t}};ce.unstable_act=Y2;ce.useCallback=function(e,t){return Dt.current.useCallback(e,t)};ce.useContext=function(e){return Dt.current.useContext(e)};ce.useDebugValue=function(){};ce.useDeferredValue=function(e){return Dt.current.useDeferredValue(e)};ce.useEffect=function(e,t){return Dt.current.useEffect(e,t)};ce.useId=function(){return Dt.current.useId()};ce.useImperativeHandle=function(e,t,n){return Dt.current.useImperativeHandle(e,t,n)};ce.useInsertionEffect=function(e,t){return Dt.current.useInsertionEffect(e,t)};ce.useLayoutEffect=function(e,t){return Dt.current.useLayoutEffect(e,t)};ce.useMemo=function(e,t){return Dt.current.useMemo(e,t)};ce.useReducer=function(e,t,n){return Dt.current.useReducer(e,t,n)};ce.useRef=function(e){return Dt.current.useRef(e)};ce.useState=function(e){return Dt.current.useState(e)};ce.useSyncExternalStore=function(e,t,n){return Dt.current.useSyncExternalStore(e,t,n)};ce.useTransition=function(){return Dt.current.useTransition()};ce.version="18.3.1";H2.exports=ce;var k=H2.exports;const E=_e(k),Q2=cO({__proto__:null,default:E},[k]);/**
 * @license React
 * react-jsx-runtime.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var CO=k,_O=Symbol.for("react.element"),AO=Symbol.for("react.fragment"),EO=Object.prototype.hasOwnProperty,TO=CO.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.ReactCurrentOwner,jO={key:!0,ref:!0,__self:!0,__source:!0};function Z2(e,t,n){var r,a={},o=null,i=null;n!==void 0&&(o=""+n),t.key!==void 0&&(o=""+t.key),t.ref!==void 0&&(i=t.ref);for(r in t)EO.call(t,r)&&!jO.hasOwnProperty(r)&&(a[r]=t[r]);if(e&&e.defaultProps)for(r in t=e.defaultProps,t)a[r]===void 0&&(a[r]=t[r]);return{$$typeof:_O,type:e,key:o,ref:i,props:a,_owner:TO.current}}yf.Fragment=AO;yf.jsx=Z2;yf.jsxs=Z2;z2.exports=yf;var b=z2.exports,Ic={},J2={exports:{}},an={},e5={exports:{}},t5={};/**
 * @license React
 * scheduler.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */(function(e){function t(N,D){var H=N.length;N.push(D);e:for(;0<H;){var W=H-1>>>1,G=N[W];if(0<a(G,D))N[W]=D,N[H]=G,H=W;else break e}}function n(N){return N.length===0?null:N[0]}function r(N){if(N.length===0)return null;var D=N[0],H=N.pop();if(H!==D){N[0]=H;e:for(var W=0,G=N.length,Z=G>>>1;W<Z;){var re=2*(W+1)-1,ve=N[re],be=re+1,J=N[be];if(0>a(ve,H))be<G&&0>a(J,ve)?(N[W]=J,N[be]=H,W=be):(N[W]=ve,N[re]=H,W=re);else if(be<G&&0>a(J,H))N[W]=J,N[be]=H,W=be;else break e}}return D}function a(N,D){var H=N.sortIndex-D.sortIndex;return H!==0?H:N.id-D.id}if(typeof performance=="object"&&typeof performance.now=="function"){var o=performance;e.unstable_now=function(){return o.now()}}else{var i=Date,s=i.now();e.unstable_now=function(){return i.now()-s}}var l=[],u=[],p=1,c=null,f=3,m=!1,h=!1,d=!1,g=typeof setTimeout=="function"?setTimeout:null,v=typeof clearTimeout=="function"?clearTimeout:null,y=typeof setImmediate<"u"?setImmediate:null;typeof navigator<"u"&&navigator.scheduling!==void 0&&navigator.scheduling.isInputPending!==void 0&&navigator.scheduling.isInputPending.bind(navigator.scheduling);function x(N){for(var D=n(u);D!==null;){if(D.callback===null)r(u);else if(D.startTime<=N)r(u),D.sortIndex=D.expirationTime,t(l,D);else break;D=n(u)}}function P(N){if(d=!1,x(N),!h)if(n(l)!==null)h=!0,R(w);else{var D=n(u);D!==null&&L(P,D.startTime-N)}}function w(N,D){h=!1,d&&(d=!1,v(C),C=-1),m=!0;var H=f;try{for(x(D),c=n(l);c!==null&&(!(c.expirationTime>D)||N&&!A());){var W=c.callback;if(typeof W=="function"){c.callback=null,f=c.priorityLevel;var G=W(c.expirationTime<=D);D=e.unstable_now(),typeof G=="function"?c.callback=G:c===n(l)&&r(l),x(D)}else r(l);c=n(l)}if(c!==null)var Z=!0;else{var re=n(u);re!==null&&L(P,re.startTime-D),Z=!1}return Z}finally{c=null,f=H,m=!1}}var S=!1,O=null,C=-1,_=5,T=-1;function A(){return!(e.unstable_now()-T<_)}function j(){if(O!==null){var N=e.unstable_now();T=N;var D=!0;try{D=O(!0,N)}finally{D?$():(S=!1,O=null)}}else S=!1}var $;if(typeof y=="function")$=function(){y(j)};else if(typeof MessageChannel<"u"){var M=new MessageChannel,I=M.port2;M.port1.onmessage=j,$=function(){I.postMessage(null)}}else $=function(){g(j,0)};function R(N){O=N,S||(S=!0,$())}function L(N,D){C=g(function(){N(e.unstable_now())},D)}e.unstable_IdlePriority=5,e.unstable_ImmediatePriority=1,e.unstable_LowPriority=4,e.unstable_NormalPriority=3,e.unstable_Profiling=null,e.unstable_UserBlockingPriority=2,e.unstable_cancelCallback=function(N){N.callback=null},e.unstable_continueExecution=function(){h||m||(h=!0,R(w))},e.unstable_forceFrameRate=function(N){0>N||125<N?console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"):_=0<N?Math.floor(1e3/N):5},e.unstable_getCurrentPriorityLevel=function(){return f},e.unstable_getFirstCallbackNode=function(){return n(l)},e.unstable_next=function(N){switch(f){case 1:case 2:case 3:var D=3;break;default:D=f}var H=f;f=D;try{return N()}finally{f=H}},e.unstable_pauseExecution=function(){},e.unstable_requestPaint=function(){},e.unstable_runWithPriority=function(N,D){switch(N){case 1:case 2:case 3:case 4:case 5:break;default:N=3}var H=f;f=N;try{return D()}finally{f=H}},e.unstable_scheduleCallback=function(N,D,H){var W=e.unstable_now();switch(typeof H=="object"&&H!==null?(H=H.delay,H=typeof H=="number"&&0<H?W+H:W):H=W,N){case 1:var G=-1;break;case 2:G=250;break;case 5:G=1073741823;break;case 4:G=1e4;break;default:G=5e3}return G=H+G,N={id:p++,callback:D,priorityLevel:N,startTime:H,expirationTime:G,sortIndex:-1},H>W?(N.sortIndex=H,t(u,N),n(l)===null&&N===n(u)&&(d?(v(C),C=-1):d=!0,L(P,H-W))):(N.sortIndex=G,t(l,N),h||m||(h=!0,R(w))),N},e.unstable_shouldYield=A,e.unstable_wrapCallback=function(N){var D=f;return function(){var H=f;f=D;try{return N.apply(this,arguments)}finally{f=H}}}})(t5);e5.exports=t5;var NO=e5.exports;/**
 * @license React
 * react-dom.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var $O=k,rn=NO;function U(e){for(var t="https://reactjs.org/docs/error-decoder.html?invariant="+e,n=1;n<arguments.length;n++)t+="&args[]="+encodeURIComponent(arguments[n]);return"Minified React error #"+e+"; visit "+t+" for the full message or use the non-minified dev environment for full errors and additional helpful warnings."}var n5=new Set,pl={};function Oo(e,t){Si(e,t),Si(e+"Capture",t)}function Si(e,t){for(pl[e]=t,e=0;e<t.length;e++)n5.add(t[e])}var $r=!(typeof window>"u"||typeof window.document>"u"||typeof window.document.createElement>"u"),qm=Object.prototype.hasOwnProperty,MO=/^[:A-Z_a-z\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u02FF\u0370-\u037D\u037F-\u1FFF\u200C-\u200D\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD][:A-Z_a-z\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u02FF\u0370-\u037D\u037F-\u1FFF\u200C-\u200D\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD\-.0-9\u00B7\u0300-\u036F\u203F-\u2040]*$/,t1={},n1={};function RO(e){return qm.call(n1,e)?!0:qm.call(t1,e)?!1:MO.test(e)?n1[e]=!0:(t1[e]=!0,!1)}function IO(e,t,n,r){if(n!==null&&n.type===0)return!1;switch(typeof t){case"function":case"symbol":return!0;case"boolean":return r?!1:n!==null?!n.acceptsBooleans:(e=e.toLowerCase().slice(0,5),e!=="data-"&&e!=="aria-");default:return!1}}function DO(e,t,n,r){if(t===null||typeof t>"u"||IO(e,t,n,r))return!0;if(r)return!1;if(n!==null)switch(n.type){case 3:return!t;case 4:return t===!1;case 5:return isNaN(t);case 6:return isNaN(t)||1>t}return!1}function Lt(e,t,n,r,a,o,i){this.acceptsBooleans=t===2||t===3||t===4,this.attributeName=r,this.attributeNamespace=a,this.mustUseProperty=n,this.propertyName=e,this.type=t,this.sanitizeURL=o,this.removeEmptyString=i}var yt={};"children dangerouslySetInnerHTML defaultValue defaultChecked innerHTML suppressContentEditableWarning suppressHydrationWarning style".split(" ").forEach(function(e){yt[e]=new Lt(e,0,!1,e,null,!1,!1)});[["acceptCharset","accept-charset"],["className","class"],["htmlFor","for"],["httpEquiv","http-equiv"]].forEach(function(e){var t=e[0];yt[t]=new Lt(t,1,!1,e[1],null,!1,!1)});["contentEditable","draggable","spellCheck","value"].forEach(function(e){yt[e]=new Lt(e,2,!1,e.toLowerCase(),null,!1,!1)});["autoReverse","externalResourcesRequired","focusable","preserveAlpha"].forEach(function(e){yt[e]=new Lt(e,2,!1,e,null,!1,!1)});"allowFullScreen async autoFocus autoPlay controls default defer disabled disablePictureInPicture disableRemotePlayback formNoValidate hidden loop noModule noValidate open playsInline readOnly required reversed scoped seamless itemScope".split(" ").forEach(function(e){yt[e]=new Lt(e,3,!1,e.toLowerCase(),null,!1,!1)});["checked","multiple","muted","selected"].forEach(function(e){yt[e]=new Lt(e,3,!0,e,null,!1,!1)});["capture","download"].forEach(function(e){yt[e]=new Lt(e,4,!1,e,null,!1,!1)});["cols","rows","size","span"].forEach(function(e){yt[e]=new Lt(e,6,!1,e,null,!1,!1)});["rowSpan","start"].forEach(function(e){yt[e]=new Lt(e,5,!1,e.toLowerCase(),null,!1,!1)});var Ev=/[\-:]([a-z])/g;function Tv(e){return e[1].toUpperCase()}"accent-height alignment-baseline arabic-form baseline-shift cap-height clip-path clip-rule color-interpolation color-interpolation-filters color-profile color-rendering dominant-baseline enable-background fill-opacity fill-rule flood-color flood-opacity font-family font-size font-size-adjust font-stretch font-style font-variant font-weight glyph-name glyph-orientation-horizontal glyph-orientation-vertical horiz-adv-x horiz-origin-x image-rendering letter-spacing lighting-color marker-end marker-mid marker-start overline-position overline-thickness paint-order panose-1 pointer-events rendering-intent shape-rendering stop-color stop-opacity strikethrough-position strikethrough-thickness stroke-dasharray stroke-dashoffset stroke-linecap stroke-linejoin stroke-miterlimit stroke-opacity stroke-width text-anchor text-decoration text-rendering underline-position underline-thickness unicode-bidi unicode-range units-per-em v-alphabetic v-hanging v-ideographic v-mathematical vector-effect vert-adv-y vert-origin-x vert-origin-y word-spacing writing-mode xmlns:xlink x-height".split(" ").forEach(function(e){var t=e.replace(Ev,Tv);yt[t]=new Lt(t,1,!1,e,null,!1,!1)});"xlink:actuate xlink:arcrole xlink:role xlink:show xlink:title xlink:type".split(" ").forEach(function(e){var t=e.replace(Ev,Tv);yt[t]=new Lt(t,1,!1,e,"http://www.w3.org/1999/xlink",!1,!1)});["xml:base","xml:lang","xml:space"].forEach(function(e){var t=e.replace(Ev,Tv);yt[t]=new Lt(t,1,!1,e,"http://www.w3.org/XML/1998/namespace",!1,!1)});["tabIndex","crossOrigin"].forEach(function(e){yt[e]=new Lt(e,1,!1,e.toLowerCase(),null,!1,!1)});yt.xlinkHref=new Lt("xlinkHref",1,!1,"xlink:href","http://www.w3.org/1999/xlink",!0,!1);["src","href","action","formAction"].forEach(function(e){yt[e]=new Lt(e,1,!1,e.toLowerCase(),null,!0,!0)});function jv(e,t,n,r){var a=yt.hasOwnProperty(t)?yt[t]:null;(a!==null?a.type!==0:r||!(2<t.length)||t[0]!=="o"&&t[0]!=="O"||t[1]!=="n"&&t[1]!=="N")&&(DO(t,n,a,r)&&(n=null),r||a===null?RO(t)&&(n===null?e.removeAttribute(t):e.setAttribute(t,""+n)):a.mustUseProperty?e[a.propertyName]=n===null?a.type===3?!1:"":n:(t=a.attributeName,r=a.attributeNamespace,n===null?e.removeAttribute(t):(a=a.type,n=a===3||a===4&&n===!0?"":""+n,r?e.setAttributeNS(r,t,n):e.setAttribute(t,n))))}var Hr=$O.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED,Iu=Symbol.for("react.element"),Fo=Symbol.for("react.portal"),Bo=Symbol.for("react.fragment"),Nv=Symbol.for("react.strict_mode"),Vm=Symbol.for("react.profiler"),r5=Symbol.for("react.provider"),a5=Symbol.for("react.context"),$v=Symbol.for("react.forward_ref"),Km=Symbol.for("react.suspense"),Xm=Symbol.for("react.suspense_list"),Mv=Symbol.for("react.memo"),Zr=Symbol.for("react.lazy"),o5=Symbol.for("react.offscreen"),r1=Symbol.iterator;function ks(e){return e===null||typeof e!="object"?null:(e=r1&&e[r1]||e["@@iterator"],typeof e=="function"?e:null)}var Ue=Object.assign,Rd;function Ws(e){if(Rd===void 0)try{throw Error()}catch(n){var t=n.stack.trim().match(/\n( *(at )?)/);Rd=t&&t[1]||""}return`
`+Rd+e}var Id=!1;function Dd(e,t){if(!e||Id)return"";Id=!0;var n=Error.prepareStackTrace;Error.prepareStackTrace=void 0;try{if(t)if(t=function(){throw Error()},Object.defineProperty(t.prototype,"props",{set:function(){throw Error()}}),typeof Reflect=="object"&&Reflect.construct){try{Reflect.construct(t,[])}catch(u){var r=u}Reflect.construct(e,[],t)}else{try{t.call()}catch(u){r=u}e.call(t.prototype)}else{try{throw Error()}catch(u){r=u}e()}}catch(u){if(u&&r&&typeof u.stack=="string"){for(var a=u.stack.split(`
`),o=r.stack.split(`
`),i=a.length-1,s=o.length-1;1<=i&&0<=s&&a[i]!==o[s];)s--;for(;1<=i&&0<=s;i--,s--)if(a[i]!==o[s]){if(i!==1||s!==1)do if(i--,s--,0>s||a[i]!==o[s]){var l=`
`+a[i].replace(" at new "," at ");return e.displayName&&l.includes("<anonymous>")&&(l=l.replace("<anonymous>",e.displayName)),l}while(1<=i&&0<=s);break}}}finally{Id=!1,Error.prepareStackTrace=n}return(e=e?e.displayName||e.name:"")?Ws(e):""}function LO(e){switch(e.tag){case 5:return Ws(e.type);case 16:return Ws("Lazy");case 13:return Ws("Suspense");case 19:return Ws("SuspenseList");case 0:case 2:case 15:return e=Dd(e.type,!1),e;case 11:return e=Dd(e.type.render,!1),e;case 1:return e=Dd(e.type,!0),e;default:return""}}function Ym(e){if(e==null)return null;if(typeof e=="function")return e.displayName||e.name||null;if(typeof e=="string")return e;switch(e){case Bo:return"Fragment";case Fo:return"Portal";case Vm:return"Profiler";case Nv:return"StrictMode";case Km:return"Suspense";case Xm:return"SuspenseList"}if(typeof e=="object")switch(e.$$typeof){case a5:return(e.displayName||"Context")+".Consumer";case r5:return(e._context.displayName||"Context")+".Provider";case $v:var t=e.render;return e=e.displayName,e||(e=t.displayName||t.name||"",e=e!==""?"ForwardRef("+e+")":"ForwardRef"),e;case Mv:return t=e.displayName||null,t!==null?t:Ym(e.type)||"Memo";case Zr:t=e._payload,e=e._init;try{return Ym(e(t))}catch{}}return null}function FO(e){var t=e.type;switch(e.tag){case 24:return"Cache";case 9:return(t.displayName||"Context")+".Consumer";case 10:return(t._context.displayName||"Context")+".Provider";case 18:return"DehydratedFragment";case 11:return e=t.render,e=e.displayName||e.name||"",t.displayName||(e!==""?"ForwardRef("+e+")":"ForwardRef");case 7:return"Fragment";case 5:return t;case 4:return"Portal";case 3:return"Root";case 6:return"Text";case 16:return Ym(t);case 8:return t===Nv?"StrictMode":"Mode";case 22:return"Offscreen";case 12:return"Profiler";case 21:return"Scope";case 13:return"Suspense";case 19:return"SuspenseList";case 25:return"TracingMarker";case 1:case 0:case 17:case 2:case 14:case 15:if(typeof t=="function")return t.displayName||t.name||null;if(typeof t=="string")return t}return null}function Pa(e){switch(typeof e){case"boolean":case"number":case"string":case"undefined":return e;case"object":return e;default:return""}}function i5(e){var t=e.type;return(e=e.nodeName)&&e.toLowerCase()==="input"&&(t==="checkbox"||t==="radio")}function BO(e){var t=i5(e)?"checked":"value",n=Object.getOwnPropertyDescriptor(e.constructor.prototype,t),r=""+e[t];if(!e.hasOwnProperty(t)&&typeof n<"u"&&typeof n.get=="function"&&typeof n.set=="function"){var a=n.get,o=n.set;return Object.defineProperty(e,t,{configurable:!0,get:function(){return a.call(this)},set:function(i){r=""+i,o.call(this,i)}}),Object.defineProperty(e,t,{enumerable:n.enumerable}),{getValue:function(){return r},setValue:function(i){r=""+i},stopTracking:function(){e._valueTracker=null,delete e[t]}}}}function Du(e){e._valueTracker||(e._valueTracker=BO(e))}function s5(e){if(!e)return!1;var t=e._valueTracker;if(!t)return!0;var n=t.getValue(),r="";return e&&(r=i5(e)?e.checked?"true":"false":e.value),e=r,e!==n?(t.setValue(e),!0):!1}function Dc(e){if(e=e||(typeof document<"u"?document:void 0),typeof e>"u")return null;try{return e.activeElement||e.body}catch{return e.body}}function Qm(e,t){var n=t.checked;return Ue({},t,{defaultChecked:void 0,defaultValue:void 0,value:void 0,checked:n??e._wrapperState.initialChecked})}function a1(e,t){var n=t.defaultValue==null?"":t.defaultValue,r=t.checked!=null?t.checked:t.defaultChecked;n=Pa(t.value!=null?t.value:n),e._wrapperState={initialChecked:r,initialValue:n,controlled:t.type==="checkbox"||t.type==="radio"?t.checked!=null:t.value!=null}}function l5(e,t){t=t.checked,t!=null&&jv(e,"checked",t,!1)}function Zm(e,t){l5(e,t);var n=Pa(t.value),r=t.type;if(n!=null)r==="number"?(n===0&&e.value===""||e.value!=n)&&(e.value=""+n):e.value!==""+n&&(e.value=""+n);else if(r==="submit"||r==="reset"){e.removeAttribute("value");return}t.hasOwnProperty("value")?Jm(e,t.type,n):t.hasOwnProperty("defaultValue")&&Jm(e,t.type,Pa(t.defaultValue)),t.checked==null&&t.defaultChecked!=null&&(e.defaultChecked=!!t.defaultChecked)}function o1(e,t,n){if(t.hasOwnProperty("value")||t.hasOwnProperty("defaultValue")){var r=t.type;if(!(r!=="submit"&&r!=="reset"||t.value!==void 0&&t.value!==null))return;t=""+e._wrapperState.initialValue,n||t===e.value||(e.value=t),e.defaultValue=t}n=e.name,n!==""&&(e.name=""),e.defaultChecked=!!e._wrapperState.initialChecked,n!==""&&(e.name=n)}function Jm(e,t,n){(t!=="number"||Dc(e.ownerDocument)!==e)&&(n==null?e.defaultValue=""+e._wrapperState.initialValue:e.defaultValue!==""+n&&(e.defaultValue=""+n))}var qs=Array.isArray;function ei(e,t,n,r){if(e=e.options,t){t={};for(var a=0;a<n.length;a++)t["$"+n[a]]=!0;for(n=0;n<e.length;n++)a=t.hasOwnProperty("$"+e[n].value),e[n].selected!==a&&(e[n].selected=a),a&&r&&(e[n].defaultSelected=!0)}else{for(n=""+Pa(n),t=null,a=0;a<e.length;a++){if(e[a].value===n){e[a].selected=!0,r&&(e[a].defaultSelected=!0);return}t!==null||e[a].disabled||(t=e[a])}t!==null&&(t.selected=!0)}}function eh(e,t){if(t.dangerouslySetInnerHTML!=null)throw Error(U(91));return Ue({},t,{value:void 0,defaultValue:void 0,children:""+e._wrapperState.initialValue})}function i1(e,t){var n=t.value;if(n==null){if(n=t.children,t=t.defaultValue,n!=null){if(t!=null)throw Error(U(92));if(qs(n)){if(1<n.length)throw Error(U(93));n=n[0]}t=n}t==null&&(t=""),n=t}e._wrapperState={initialValue:Pa(n)}}function u5(e,t){var n=Pa(t.value),r=Pa(t.defaultValue);n!=null&&(n=""+n,n!==e.value&&(e.value=n),t.defaultValue==null&&e.defaultValue!==n&&(e.defaultValue=n)),r!=null&&(e.defaultValue=""+r)}function s1(e){var t=e.textContent;t===e._wrapperState.initialValue&&t!==""&&t!==null&&(e.value=t)}function c5(e){switch(e){case"svg":return"http://www.w3.org/2000/svg";case"math":return"http://www.w3.org/1998/Math/MathML";default:return"http://www.w3.org/1999/xhtml"}}function th(e,t){return e==null||e==="http://www.w3.org/1999/xhtml"?c5(t):e==="http://www.w3.org/2000/svg"&&t==="foreignObject"?"http://www.w3.org/1999/xhtml":e}var Lu,p5=function(e){return typeof MSApp<"u"&&MSApp.execUnsafeLocalFunction?function(t,n,r,a){MSApp.execUnsafeLocalFunction(function(){return e(t,n,r,a)})}:e}(function(e,t){if(e.namespaceURI!=="http://www.w3.org/2000/svg"||"innerHTML"in e)e.innerHTML=t;else{for(Lu=Lu||document.createElement("div"),Lu.innerHTML="<svg>"+t.valueOf().toString()+"</svg>",t=Lu.firstChild;e.firstChild;)e.removeChild(e.firstChild);for(;t.firstChild;)e.appendChild(t.firstChild)}});function fl(e,t){if(t){var n=e.firstChild;if(n&&n===e.lastChild&&n.nodeType===3){n.nodeValue=t;return}}e.textContent=t}var Qs={animationIterationCount:!0,aspectRatio:!0,borderImageOutset:!0,borderImageSlice:!0,borderImageWidth:!0,boxFlex:!0,boxFlexGroup:!0,boxOrdinalGroup:!0,columnCount:!0,columns:!0,flex:!0,flexGrow:!0,flexPositive:!0,flexShrink:!0,flexNegative:!0,flexOrder:!0,gridArea:!0,gridRow:!0,gridRowEnd:!0,gridRowSpan:!0,gridRowStart:!0,gridColumn:!0,gridColumnEnd:!0,gridColumnSpan:!0,gridColumnStart:!0,fontWeight:!0,lineClamp:!0,lineHeight:!0,opacity:!0,order:!0,orphans:!0,tabSize:!0,widows:!0,zIndex:!0,zoom:!0,fillOpacity:!0,floodOpacity:!0,stopOpacity:!0,strokeDasharray:!0,strokeDashoffset:!0,strokeMiterlimit:!0,strokeOpacity:!0,strokeWidth:!0},zO=["Webkit","ms","Moz","O"];Object.keys(Qs).forEach(function(e){zO.forEach(function(t){t=t+e.charAt(0).toUpperCase()+e.substring(1),Qs[t]=Qs[e]})});function f5(e,t,n){return t==null||typeof t=="boolean"||t===""?"":n||typeof t!="number"||t===0||Qs.hasOwnProperty(e)&&Qs[e]?(""+t).trim():t+"px"}function d5(e,t){e=e.style;for(var n in t)if(t.hasOwnProperty(n)){var r=n.indexOf("--")===0,a=f5(n,t[n],r);n==="float"&&(n="cssFloat"),r?e.setProperty(n,a):e[n]=a}}var HO=Ue({menuitem:!0},{area:!0,base:!0,br:!0,col:!0,embed:!0,hr:!0,img:!0,input:!0,keygen:!0,link:!0,meta:!0,param:!0,source:!0,track:!0,wbr:!0});function nh(e,t){if(t){if(HO[e]&&(t.children!=null||t.dangerouslySetInnerHTML!=null))throw Error(U(137,e));if(t.dangerouslySetInnerHTML!=null){if(t.children!=null)throw Error(U(60));if(typeof t.dangerouslySetInnerHTML!="object"||!("__html"in t.dangerouslySetInnerHTML))throw Error(U(61))}if(t.style!=null&&typeof t.style!="object")throw Error(U(62))}}function rh(e,t){if(e.indexOf("-")===-1)return typeof t.is=="string";switch(e){case"annotation-xml":case"color-profile":case"font-face":case"font-face-src":case"font-face-uri":case"font-face-format":case"font-face-name":case"missing-glyph":return!1;default:return!0}}var ah=null;function Rv(e){return e=e.target||e.srcElement||window,e.correspondingUseElement&&(e=e.correspondingUseElement),e.nodeType===3?e.parentNode:e}var oh=null,ti=null,ni=null;function l1(e){if(e=wu(e)){if(typeof oh!="function")throw Error(U(280));var t=e.stateNode;t&&(t=Pf(t),oh(e.stateNode,e.type,t))}}function m5(e){ti?ni?ni.push(e):ni=[e]:ti=e}function h5(){if(ti){var e=ti,t=ni;if(ni=ti=null,l1(e),t)for(e=0;e<t.length;e++)l1(t[e])}}function v5(e,t){return e(t)}function y5(){}var Ld=!1;function g5(e,t,n){if(Ld)return e(t,n);Ld=!0;try{return v5(e,t,n)}finally{Ld=!1,(ti!==null||ni!==null)&&(y5(),h5())}}function dl(e,t){var n=e.stateNode;if(n===null)return null;var r=Pf(n);if(r===null)return null;n=r[t];e:switch(t){case"onClick":case"onClickCapture":case"onDoubleClick":case"onDoubleClickCapture":case"onMouseDown":case"onMouseDownCapture":case"onMouseMove":case"onMouseMoveCapture":case"onMouseUp":case"onMouseUpCapture":case"onMouseEnter":(r=!r.disabled)||(e=e.type,r=!(e==="button"||e==="input"||e==="select"||e==="textarea")),e=!r;break e;default:e=!1}if(e)return null;if(n&&typeof n!="function")throw Error(U(231,t,typeof n));return n}var ih=!1;if($r)try{var Cs={};Object.defineProperty(Cs,"passive",{get:function(){ih=!0}}),window.addEventListener("test",Cs,Cs),window.removeEventListener("test",Cs,Cs)}catch{ih=!1}function GO(e,t,n,r,a,o,i,s,l){var u=Array.prototype.slice.call(arguments,3);try{t.apply(n,u)}catch(p){this.onError(p)}}var Zs=!1,Lc=null,Fc=!1,sh=null,UO={onError:function(e){Zs=!0,Lc=e}};function WO(e,t,n,r,a,o,i,s,l){Zs=!1,Lc=null,GO.apply(UO,arguments)}function qO(e,t,n,r,a,o,i,s,l){if(WO.apply(this,arguments),Zs){if(Zs){var u=Lc;Zs=!1,Lc=null}else throw Error(U(198));Fc||(Fc=!0,sh=u)}}function ko(e){var t=e,n=e;if(e.alternate)for(;t.return;)t=t.return;else{e=t;do t=e,t.flags&4098&&(n=t.return),e=t.return;while(e)}return t.tag===3?n:null}function x5(e){if(e.tag===13){var t=e.memoizedState;if(t===null&&(e=e.alternate,e!==null&&(t=e.memoizedState)),t!==null)return t.dehydrated}return null}function u1(e){if(ko(e)!==e)throw Error(U(188))}function VO(e){var t=e.alternate;if(!t){if(t=ko(e),t===null)throw Error(U(188));return t!==e?null:e}for(var n=e,r=t;;){var a=n.return;if(a===null)break;var o=a.alternate;if(o===null){if(r=a.return,r!==null){n=r;continue}break}if(a.child===o.child){for(o=a.child;o;){if(o===n)return u1(a),e;if(o===r)return u1(a),t;o=o.sibling}throw Error(U(188))}if(n.return!==r.return)n=a,r=o;else{for(var i=!1,s=a.child;s;){if(s===n){i=!0,n=a,r=o;break}if(s===r){i=!0,r=a,n=o;break}s=s.sibling}if(!i){for(s=o.child;s;){if(s===n){i=!0,n=o,r=a;break}if(s===r){i=!0,r=o,n=a;break}s=s.sibling}if(!i)throw Error(U(189))}}if(n.alternate!==r)throw Error(U(190))}if(n.tag!==3)throw Error(U(188));return n.stateNode.current===n?e:t}function w5(e){return e=VO(e),e!==null?b5(e):null}function b5(e){if(e.tag===5||e.tag===6)return e;for(e=e.child;e!==null;){var t=b5(e);if(t!==null)return t;e=e.sibling}return null}var P5=rn.unstable_scheduleCallback,c1=rn.unstable_cancelCallback,KO=rn.unstable_shouldYield,XO=rn.unstable_requestPaint,Ke=rn.unstable_now,YO=rn.unstable_getCurrentPriorityLevel,Iv=rn.unstable_ImmediatePriority,S5=rn.unstable_UserBlockingPriority,Bc=rn.unstable_NormalPriority,QO=rn.unstable_LowPriority,O5=rn.unstable_IdlePriority,gf=null,sr=null;function ZO(e){if(sr&&typeof sr.onCommitFiberRoot=="function")try{sr.onCommitFiberRoot(gf,e,void 0,(e.current.flags&128)===128)}catch{}}var zn=Math.clz32?Math.clz32:tk,JO=Math.log,ek=Math.LN2;function tk(e){return e>>>=0,e===0?32:31-(JO(e)/ek|0)|0}var Fu=64,Bu=4194304;function Vs(e){switch(e&-e){case 1:return 1;case 2:return 2;case 4:return 4;case 8:return 8;case 16:return 16;case 32:return 32;case 64:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:return e&4194240;case 4194304:case 8388608:case 16777216:case 33554432:case 67108864:return e&130023424;case 134217728:return 134217728;case 268435456:return 268435456;case 536870912:return 536870912;case 1073741824:return 1073741824;default:return e}}function zc(e,t){var n=e.pendingLanes;if(n===0)return 0;var r=0,a=e.suspendedLanes,o=e.pingedLanes,i=n&268435455;if(i!==0){var s=i&~a;s!==0?r=Vs(s):(o&=i,o!==0&&(r=Vs(o)))}else i=n&~a,i!==0?r=Vs(i):o!==0&&(r=Vs(o));if(r===0)return 0;if(t!==0&&t!==r&&!(t&a)&&(a=r&-r,o=t&-t,a>=o||a===16&&(o&4194240)!==0))return t;if(r&4&&(r|=n&16),t=e.entangledLanes,t!==0)for(e=e.entanglements,t&=r;0<t;)n=31-zn(t),a=1<<n,r|=e[n],t&=~a;return r}function nk(e,t){switch(e){case 1:case 2:case 4:return t+250;case 8:case 16:case 32:case 64:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:return t+5e3;case 4194304:case 8388608:case 16777216:case 33554432:case 67108864:return-1;case 134217728:case 268435456:case 536870912:case 1073741824:return-1;default:return-1}}function rk(e,t){for(var n=e.suspendedLanes,r=e.pingedLanes,a=e.expirationTimes,o=e.pendingLanes;0<o;){var i=31-zn(o),s=1<<i,l=a[i];l===-1?(!(s&n)||s&r)&&(a[i]=nk(s,t)):l<=t&&(e.expiredLanes|=s),o&=~s}}function lh(e){return e=e.pendingLanes&-1073741825,e!==0?e:e&1073741824?1073741824:0}function k5(){var e=Fu;return Fu<<=1,!(Fu&4194240)&&(Fu=64),e}function Fd(e){for(var t=[],n=0;31>n;n++)t.push(e);return t}function gu(e,t,n){e.pendingLanes|=t,t!==536870912&&(e.suspendedLanes=0,e.pingedLanes=0),e=e.eventTimes,t=31-zn(t),e[t]=n}function ak(e,t){var n=e.pendingLanes&~t;e.pendingLanes=t,e.suspendedLanes=0,e.pingedLanes=0,e.expiredLanes&=t,e.mutableReadLanes&=t,e.entangledLanes&=t,t=e.entanglements;var r=e.eventTimes;for(e=e.expirationTimes;0<n;){var a=31-zn(n),o=1<<a;t[a]=0,r[a]=-1,e[a]=-1,n&=~o}}function Dv(e,t){var n=e.entangledLanes|=t;for(e=e.entanglements;n;){var r=31-zn(n),a=1<<r;a&t|e[r]&t&&(e[r]|=t),n&=~a}}var Oe=0;function C5(e){return e&=-e,1<e?4<e?e&268435455?16:536870912:4:1}var _5,Lv,A5,E5,T5,uh=!1,zu=[],da=null,ma=null,ha=null,ml=new Map,hl=new Map,ta=[],ok="mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset submit".split(" ");function p1(e,t){switch(e){case"focusin":case"focusout":da=null;break;case"dragenter":case"dragleave":ma=null;break;case"mouseover":case"mouseout":ha=null;break;case"pointerover":case"pointerout":ml.delete(t.pointerId);break;case"gotpointercapture":case"lostpointercapture":hl.delete(t.pointerId)}}function _s(e,t,n,r,a,o){return e===null||e.nativeEvent!==o?(e={blockedOn:t,domEventName:n,eventSystemFlags:r,nativeEvent:o,targetContainers:[a]},t!==null&&(t=wu(t),t!==null&&Lv(t)),e):(e.eventSystemFlags|=r,t=e.targetContainers,a!==null&&t.indexOf(a)===-1&&t.push(a),e)}function ik(e,t,n,r,a){switch(t){case"focusin":return da=_s(da,e,t,n,r,a),!0;case"dragenter":return ma=_s(ma,e,t,n,r,a),!0;case"mouseover":return ha=_s(ha,e,t,n,r,a),!0;case"pointerover":var o=a.pointerId;return ml.set(o,_s(ml.get(o)||null,e,t,n,r,a)),!0;case"gotpointercapture":return o=a.pointerId,hl.set(o,_s(hl.get(o)||null,e,t,n,r,a)),!0}return!1}function j5(e){var t=qa(e.target);if(t!==null){var n=ko(t);if(n!==null){if(t=n.tag,t===13){if(t=x5(n),t!==null){e.blockedOn=t,T5(e.priority,function(){A5(n)});return}}else if(t===3&&n.stateNode.current.memoizedState.isDehydrated){e.blockedOn=n.tag===3?n.stateNode.containerInfo:null;return}}}e.blockedOn=null}function bc(e){if(e.blockedOn!==null)return!1;for(var t=e.targetContainers;0<t.length;){var n=ch(e.domEventName,e.eventSystemFlags,t[0],e.nativeEvent);if(n===null){n=e.nativeEvent;var r=new n.constructor(n.type,n);ah=r,n.target.dispatchEvent(r),ah=null}else return t=wu(n),t!==null&&Lv(t),e.blockedOn=n,!1;t.shift()}return!0}function f1(e,t,n){bc(e)&&n.delete(t)}function sk(){uh=!1,da!==null&&bc(da)&&(da=null),ma!==null&&bc(ma)&&(ma=null),ha!==null&&bc(ha)&&(ha=null),ml.forEach(f1),hl.forEach(f1)}function As(e,t){e.blockedOn===t&&(e.blockedOn=null,uh||(uh=!0,rn.unstable_scheduleCallback(rn.unstable_NormalPriority,sk)))}function vl(e){function t(a){return As(a,e)}if(0<zu.length){As(zu[0],e);for(var n=1;n<zu.length;n++){var r=zu[n];r.blockedOn===e&&(r.blockedOn=null)}}for(da!==null&&As(da,e),ma!==null&&As(ma,e),ha!==null&&As(ha,e),ml.forEach(t),hl.forEach(t),n=0;n<ta.length;n++)r=ta[n],r.blockedOn===e&&(r.blockedOn=null);for(;0<ta.length&&(n=ta[0],n.blockedOn===null);)j5(n),n.blockedOn===null&&ta.shift()}var ri=Hr.ReactCurrentBatchConfig,Hc=!0;function lk(e,t,n,r){var a=Oe,o=ri.transition;ri.transition=null;try{Oe=1,Fv(e,t,n,r)}finally{Oe=a,ri.transition=o}}function uk(e,t,n,r){var a=Oe,o=ri.transition;ri.transition=null;try{Oe=4,Fv(e,t,n,r)}finally{Oe=a,ri.transition=o}}function Fv(e,t,n,r){if(Hc){var a=ch(e,t,n,r);if(a===null)Xd(e,t,r,Gc,n),p1(e,r);else if(ik(a,e,t,n,r))r.stopPropagation();else if(p1(e,r),t&4&&-1<ok.indexOf(e)){for(;a!==null;){var o=wu(a);if(o!==null&&_5(o),o=ch(e,t,n,r),o===null&&Xd(e,t,r,Gc,n),o===a)break;a=o}a!==null&&r.stopPropagation()}else Xd(e,t,r,null,n)}}var Gc=null;function ch(e,t,n,r){if(Gc=null,e=Rv(r),e=qa(e),e!==null)if(t=ko(e),t===null)e=null;else if(n=t.tag,n===13){if(e=x5(t),e!==null)return e;e=null}else if(n===3){if(t.stateNode.current.memoizedState.isDehydrated)return t.tag===3?t.stateNode.containerInfo:null;e=null}else t!==e&&(e=null);return Gc=e,null}function N5(e){switch(e){case"cancel":case"click":case"close":case"contextmenu":case"copy":case"cut":case"auxclick":case"dblclick":case"dragend":case"dragstart":case"drop":case"focusin":case"focusout":case"input":case"invalid":case"keydown":case"keypress":case"keyup":case"mousedown":case"mouseup":case"paste":case"pause":case"play":case"pointercancel":case"pointerdown":case"pointerup":case"ratechange":case"reset":case"resize":case"seeked":case"submit":case"touchcancel":case"touchend":case"touchstart":case"volumechange":case"change":case"selectionchange":case"textInput":case"compositionstart":case"compositionend":case"compositionupdate":case"beforeblur":case"afterblur":case"beforeinput":case"blur":case"fullscreenchange":case"focus":case"hashchange":case"popstate":case"select":case"selectstart":return 1;case"drag":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"mousemove":case"mouseout":case"mouseover":case"pointermove":case"pointerout":case"pointerover":case"scroll":case"toggle":case"touchmove":case"wheel":case"mouseenter":case"mouseleave":case"pointerenter":case"pointerleave":return 4;case"message":switch(YO()){case Iv:return 1;case S5:return 4;case Bc:case QO:return 16;case O5:return 536870912;default:return 16}default:return 16}}var ua=null,Bv=null,Pc=null;function $5(){if(Pc)return Pc;var e,t=Bv,n=t.length,r,a="value"in ua?ua.value:ua.textContent,o=a.length;for(e=0;e<n&&t[e]===a[e];e++);var i=n-e;for(r=1;r<=i&&t[n-r]===a[o-r];r++);return Pc=a.slice(e,1<r?1-r:void 0)}function Sc(e){var t=e.keyCode;return"charCode"in e?(e=e.charCode,e===0&&t===13&&(e=13)):e=t,e===10&&(e=13),32<=e||e===13?e:0}function Hu(){return!0}function d1(){return!1}function on(e){function t(n,r,a,o,i){this._reactName=n,this._targetInst=a,this.type=r,this.nativeEvent=o,this.target=i,this.currentTarget=null;for(var s in e)e.hasOwnProperty(s)&&(n=e[s],this[s]=n?n(o):o[s]);return this.isDefaultPrevented=(o.defaultPrevented!=null?o.defaultPrevented:o.returnValue===!1)?Hu:d1,this.isPropagationStopped=d1,this}return Ue(t.prototype,{preventDefault:function(){this.defaultPrevented=!0;var n=this.nativeEvent;n&&(n.preventDefault?n.preventDefault():typeof n.returnValue!="unknown"&&(n.returnValue=!1),this.isDefaultPrevented=Hu)},stopPropagation:function(){var n=this.nativeEvent;n&&(n.stopPropagation?n.stopPropagation():typeof n.cancelBubble!="unknown"&&(n.cancelBubble=!0),this.isPropagationStopped=Hu)},persist:function(){},isPersistent:Hu}),t}var rs={eventPhase:0,bubbles:0,cancelable:0,timeStamp:function(e){return e.timeStamp||Date.now()},defaultPrevented:0,isTrusted:0},zv=on(rs),xu=Ue({},rs,{view:0,detail:0}),ck=on(xu),Bd,zd,Es,xf=Ue({},xu,{screenX:0,screenY:0,clientX:0,clientY:0,pageX:0,pageY:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,getModifierState:Hv,button:0,buttons:0,relatedTarget:function(e){return e.relatedTarget===void 0?e.fromElement===e.srcElement?e.toElement:e.fromElement:e.relatedTarget},movementX:function(e){return"movementX"in e?e.movementX:(e!==Es&&(Es&&e.type==="mousemove"?(Bd=e.screenX-Es.screenX,zd=e.screenY-Es.screenY):zd=Bd=0,Es=e),Bd)},movementY:function(e){return"movementY"in e?e.movementY:zd}}),m1=on(xf),pk=Ue({},xf,{dataTransfer:0}),fk=on(pk),dk=Ue({},xu,{relatedTarget:0}),Hd=on(dk),mk=Ue({},rs,{animationName:0,elapsedTime:0,pseudoElement:0}),hk=on(mk),vk=Ue({},rs,{clipboardData:function(e){return"clipboardData"in e?e.clipboardData:window.clipboardData}}),yk=on(vk),gk=Ue({},rs,{data:0}),h1=on(gk),xk={Esc:"Escape",Spacebar:" ",Left:"ArrowLeft",Up:"ArrowUp",Right:"ArrowRight",Down:"ArrowDown",Del:"Delete",Win:"OS",Menu:"ContextMenu",Apps:"ContextMenu",Scroll:"ScrollLock",MozPrintableKey:"Unidentified"},wk={8:"Backspace",9:"Tab",12:"Clear",13:"Enter",16:"Shift",17:"Control",18:"Alt",19:"Pause",20:"CapsLock",27:"Escape",32:" ",33:"PageUp",34:"PageDown",35:"End",36:"Home",37:"ArrowLeft",38:"ArrowUp",39:"ArrowRight",40:"ArrowDown",45:"Insert",46:"Delete",112:"F1",113:"F2",114:"F3",115:"F4",116:"F5",117:"F6",118:"F7",119:"F8",120:"F9",121:"F10",122:"F11",123:"F12",144:"NumLock",145:"ScrollLock",224:"Meta"},bk={Alt:"altKey",Control:"ctrlKey",Meta:"metaKey",Shift:"shiftKey"};function Pk(e){var t=this.nativeEvent;return t.getModifierState?t.getModifierState(e):(e=bk[e])?!!t[e]:!1}function Hv(){return Pk}var Sk=Ue({},xu,{key:function(e){if(e.key){var t=xk[e.key]||e.key;if(t!=="Unidentified")return t}return e.type==="keypress"?(e=Sc(e),e===13?"Enter":String.fromCharCode(e)):e.type==="keydown"||e.type==="keyup"?wk[e.keyCode]||"Unidentified":""},code:0,location:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,repeat:0,locale:0,getModifierState:Hv,charCode:function(e){return e.type==="keypress"?Sc(e):0},keyCode:function(e){return e.type==="keydown"||e.type==="keyup"?e.keyCode:0},which:function(e){return e.type==="keypress"?Sc(e):e.type==="keydown"||e.type==="keyup"?e.keyCode:0}}),Ok=on(Sk),kk=Ue({},xf,{pointerId:0,width:0,height:0,pressure:0,tangentialPressure:0,tiltX:0,tiltY:0,twist:0,pointerType:0,isPrimary:0}),v1=on(kk),Ck=Ue({},xu,{touches:0,targetTouches:0,changedTouches:0,altKey:0,metaKey:0,ctrlKey:0,shiftKey:0,getModifierState:Hv}),_k=on(Ck),Ak=Ue({},rs,{propertyName:0,elapsedTime:0,pseudoElement:0}),Ek=on(Ak),Tk=Ue({},xf,{deltaX:function(e){return"deltaX"in e?e.deltaX:"wheelDeltaX"in e?-e.wheelDeltaX:0},deltaY:function(e){return"deltaY"in e?e.deltaY:"wheelDeltaY"in e?-e.wheelDeltaY:"wheelDelta"in e?-e.wheelDelta:0},deltaZ:0,deltaMode:0}),jk=on(Tk),Nk=[9,13,27,32],Gv=$r&&"CompositionEvent"in window,Js=null;$r&&"documentMode"in document&&(Js=document.documentMode);var $k=$r&&"TextEvent"in window&&!Js,M5=$r&&(!Gv||Js&&8<Js&&11>=Js),y1=" ",g1=!1;function R5(e,t){switch(e){case"keyup":return Nk.indexOf(t.keyCode)!==-1;case"keydown":return t.keyCode!==229;case"keypress":case"mousedown":case"focusout":return!0;default:return!1}}function I5(e){return e=e.detail,typeof e=="object"&&"data"in e?e.data:null}var zo=!1;function Mk(e,t){switch(e){case"compositionend":return I5(t);case"keypress":return t.which!==32?null:(g1=!0,y1);case"textInput":return e=t.data,e===y1&&g1?null:e;default:return null}}function Rk(e,t){if(zo)return e==="compositionend"||!Gv&&R5(e,t)?(e=$5(),Pc=Bv=ua=null,zo=!1,e):null;switch(e){case"paste":return null;case"keypress":if(!(t.ctrlKey||t.altKey||t.metaKey)||t.ctrlKey&&t.altKey){if(t.char&&1<t.char.length)return t.char;if(t.which)return String.fromCharCode(t.which)}return null;case"compositionend":return M5&&t.locale!=="ko"?null:t.data;default:return null}}var Ik={color:!0,date:!0,datetime:!0,"datetime-local":!0,email:!0,month:!0,number:!0,password:!0,range:!0,search:!0,tel:!0,text:!0,time:!0,url:!0,week:!0};function x1(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t==="input"?!!Ik[e.type]:t==="textarea"}function D5(e,t,n,r){m5(r),t=Uc(t,"onChange"),0<t.length&&(n=new zv("onChange","change",null,n,r),e.push({event:n,listeners:t}))}var el=null,yl=null;function Dk(e){K5(e,0)}function wf(e){var t=Uo(e);if(s5(t))return e}function Lk(e,t){if(e==="change")return t}var L5=!1;if($r){var Gd;if($r){var Ud="oninput"in document;if(!Ud){var w1=document.createElement("div");w1.setAttribute("oninput","return;"),Ud=typeof w1.oninput=="function"}Gd=Ud}else Gd=!1;L5=Gd&&(!document.documentMode||9<document.documentMode)}function b1(){el&&(el.detachEvent("onpropertychange",F5),yl=el=null)}function F5(e){if(e.propertyName==="value"&&wf(yl)){var t=[];D5(t,yl,e,Rv(e)),g5(Dk,t)}}function Fk(e,t,n){e==="focusin"?(b1(),el=t,yl=n,el.attachEvent("onpropertychange",F5)):e==="focusout"&&b1()}function Bk(e){if(e==="selectionchange"||e==="keyup"||e==="keydown")return wf(yl)}function zk(e,t){if(e==="click")return wf(t)}function Hk(e,t){if(e==="input"||e==="change")return wf(t)}function Gk(e,t){return e===t&&(e!==0||1/e===1/t)||e!==e&&t!==t}var Gn=typeof Object.is=="function"?Object.is:Gk;function gl(e,t){if(Gn(e,t))return!0;if(typeof e!="object"||e===null||typeof t!="object"||t===null)return!1;var n=Object.keys(e),r=Object.keys(t);if(n.length!==r.length)return!1;for(r=0;r<n.length;r++){var a=n[r];if(!qm.call(t,a)||!Gn(e[a],t[a]))return!1}return!0}function P1(e){for(;e&&e.firstChild;)e=e.firstChild;return e}function S1(e,t){var n=P1(e);e=0;for(var r;n;){if(n.nodeType===3){if(r=e+n.textContent.length,e<=t&&r>=t)return{node:n,offset:t-e};e=r}e:{for(;n;){if(n.nextSibling){n=n.nextSibling;break e}n=n.parentNode}n=void 0}n=P1(n)}}function B5(e,t){return e&&t?e===t?!0:e&&e.nodeType===3?!1:t&&t.nodeType===3?B5(e,t.parentNode):"contains"in e?e.contains(t):e.compareDocumentPosition?!!(e.compareDocumentPosition(t)&16):!1:!1}function z5(){for(var e=window,t=Dc();t instanceof e.HTMLIFrameElement;){try{var n=typeof t.contentWindow.location.href=="string"}catch{n=!1}if(n)e=t.contentWindow;else break;t=Dc(e.document)}return t}function Uv(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t&&(t==="input"&&(e.type==="text"||e.type==="search"||e.type==="tel"||e.type==="url"||e.type==="password")||t==="textarea"||e.contentEditable==="true")}function Uk(e){var t=z5(),n=e.focusedElem,r=e.selectionRange;if(t!==n&&n&&n.ownerDocument&&B5(n.ownerDocument.documentElement,n)){if(r!==null&&Uv(n)){if(t=r.start,e=r.end,e===void 0&&(e=t),"selectionStart"in n)n.selectionStart=t,n.selectionEnd=Math.min(e,n.value.length);else if(e=(t=n.ownerDocument||document)&&t.defaultView||window,e.getSelection){e=e.getSelection();var a=n.textContent.length,o=Math.min(r.start,a);r=r.end===void 0?o:Math.min(r.end,a),!e.extend&&o>r&&(a=r,r=o,o=a),a=S1(n,o);var i=S1(n,r);a&&i&&(e.rangeCount!==1||e.anchorNode!==a.node||e.anchorOffset!==a.offset||e.focusNode!==i.node||e.focusOffset!==i.offset)&&(t=t.createRange(),t.setStart(a.node,a.offset),e.removeAllRanges(),o>r?(e.addRange(t),e.extend(i.node,i.offset)):(t.setEnd(i.node,i.offset),e.addRange(t)))}}for(t=[],e=n;e=e.parentNode;)e.nodeType===1&&t.push({element:e,left:e.scrollLeft,top:e.scrollTop});for(typeof n.focus=="function"&&n.focus(),n=0;n<t.length;n++)e=t[n],e.element.scrollLeft=e.left,e.element.scrollTop=e.top}}var Wk=$r&&"documentMode"in document&&11>=document.documentMode,Ho=null,ph=null,tl=null,fh=!1;function O1(e,t,n){var r=n.window===n?n.document:n.nodeType===9?n:n.ownerDocument;fh||Ho==null||Ho!==Dc(r)||(r=Ho,"selectionStart"in r&&Uv(r)?r={start:r.selectionStart,end:r.selectionEnd}:(r=(r.ownerDocument&&r.ownerDocument.defaultView||window).getSelection(),r={anchorNode:r.anchorNode,anchorOffset:r.anchorOffset,focusNode:r.focusNode,focusOffset:r.focusOffset}),tl&&gl(tl,r)||(tl=r,r=Uc(ph,"onSelect"),0<r.length&&(t=new zv("onSelect","select",null,t,n),e.push({event:t,listeners:r}),t.target=Ho)))}function Gu(e,t){var n={};return n[e.toLowerCase()]=t.toLowerCase(),n["Webkit"+e]="webkit"+t,n["Moz"+e]="moz"+t,n}var Go={animationend:Gu("Animation","AnimationEnd"),animationiteration:Gu("Animation","AnimationIteration"),animationstart:Gu("Animation","AnimationStart"),transitionend:Gu("Transition","TransitionEnd")},Wd={},H5={};$r&&(H5=document.createElement("div").style,"AnimationEvent"in window||(delete Go.animationend.animation,delete Go.animationiteration.animation,delete Go.animationstart.animation),"TransitionEvent"in window||delete Go.transitionend.transition);function bf(e){if(Wd[e])return Wd[e];if(!Go[e])return e;var t=Go[e],n;for(n in t)if(t.hasOwnProperty(n)&&n in H5)return Wd[e]=t[n];return e}var G5=bf("animationend"),U5=bf("animationiteration"),W5=bf("animationstart"),q5=bf("transitionend"),V5=new Map,k1="abort auxClick cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");function _a(e,t){V5.set(e,t),Oo(t,[e])}for(var qd=0;qd<k1.length;qd++){var Vd=k1[qd],qk=Vd.toLowerCase(),Vk=Vd[0].toUpperCase()+Vd.slice(1);_a(qk,"on"+Vk)}_a(G5,"onAnimationEnd");_a(U5,"onAnimationIteration");_a(W5,"onAnimationStart");_a("dblclick","onDoubleClick");_a("focusin","onFocus");_a("focusout","onBlur");_a(q5,"onTransitionEnd");Si("onMouseEnter",["mouseout","mouseover"]);Si("onMouseLeave",["mouseout","mouseover"]);Si("onPointerEnter",["pointerout","pointerover"]);Si("onPointerLeave",["pointerout","pointerover"]);Oo("onChange","change click focusin focusout input keydown keyup selectionchange".split(" "));Oo("onSelect","focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" "));Oo("onBeforeInput",["compositionend","keypress","textInput","paste"]);Oo("onCompositionEnd","compositionend focusout keydown keypress keyup mousedown".split(" "));Oo("onCompositionStart","compositionstart focusout keydown keypress keyup mousedown".split(" "));Oo("onCompositionUpdate","compositionupdate focusout keydown keypress keyup mousedown".split(" "));var Ks="abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "),Kk=new Set("cancel close invalid load scroll toggle".split(" ").concat(Ks));function C1(e,t,n){var r=e.type||"unknown-event";e.currentTarget=n,qO(r,t,void 0,e),e.currentTarget=null}function K5(e,t){t=(t&4)!==0;for(var n=0;n<e.length;n++){var r=e[n],a=r.event;r=r.listeners;e:{var o=void 0;if(t)for(var i=r.length-1;0<=i;i--){var s=r[i],l=s.instance,u=s.currentTarget;if(s=s.listener,l!==o&&a.isPropagationStopped())break e;C1(a,s,u),o=l}else for(i=0;i<r.length;i++){if(s=r[i],l=s.instance,u=s.currentTarget,s=s.listener,l!==o&&a.isPropagationStopped())break e;C1(a,s,u),o=l}}}if(Fc)throw e=sh,Fc=!1,sh=null,e}function Ie(e,t){var n=t[yh];n===void 0&&(n=t[yh]=new Set);var r=e+"__bubble";n.has(r)||(X5(t,e,2,!1),n.add(r))}function Kd(e,t,n){var r=0;t&&(r|=4),X5(n,e,r,t)}var Uu="_reactListening"+Math.random().toString(36).slice(2);function xl(e){if(!e[Uu]){e[Uu]=!0,n5.forEach(function(n){n!=="selectionchange"&&(Kk.has(n)||Kd(n,!1,e),Kd(n,!0,e))});var t=e.nodeType===9?e:e.ownerDocument;t===null||t[Uu]||(t[Uu]=!0,Kd("selectionchange",!1,t))}}function X5(e,t,n,r){switch(N5(t)){case 1:var a=lk;break;case 4:a=uk;break;default:a=Fv}n=a.bind(null,t,n,e),a=void 0,!ih||t!=="touchstart"&&t!=="touchmove"&&t!=="wheel"||(a=!0),r?a!==void 0?e.addEventListener(t,n,{capture:!0,passive:a}):e.addEventListener(t,n,!0):a!==void 0?e.addEventListener(t,n,{passive:a}):e.addEventListener(t,n,!1)}function Xd(e,t,n,r,a){var o=r;if(!(t&1)&&!(t&2)&&r!==null)e:for(;;){if(r===null)return;var i=r.tag;if(i===3||i===4){var s=r.stateNode.containerInfo;if(s===a||s.nodeType===8&&s.parentNode===a)break;if(i===4)for(i=r.return;i!==null;){var l=i.tag;if((l===3||l===4)&&(l=i.stateNode.containerInfo,l===a||l.nodeType===8&&l.parentNode===a))return;i=i.return}for(;s!==null;){if(i=qa(s),i===null)return;if(l=i.tag,l===5||l===6){r=o=i;continue e}s=s.parentNode}}r=r.return}g5(function(){var u=o,p=Rv(n),c=[];e:{var f=V5.get(e);if(f!==void 0){var m=zv,h=e;switch(e){case"keypress":if(Sc(n)===0)break e;case"keydown":case"keyup":m=Ok;break;case"focusin":h="focus",m=Hd;break;case"focusout":h="blur",m=Hd;break;case"beforeblur":case"afterblur":m=Hd;break;case"click":if(n.button===2)break e;case"auxclick":case"dblclick":case"mousedown":case"mousemove":case"mouseup":case"mouseout":case"mouseover":case"contextmenu":m=m1;break;case"drag":case"dragend":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"dragstart":case"drop":m=fk;break;case"touchcancel":case"touchend":case"touchmove":case"touchstart":m=_k;break;case G5:case U5:case W5:m=hk;break;case q5:m=Ek;break;case"scroll":m=ck;break;case"wheel":m=jk;break;case"copy":case"cut":case"paste":m=yk;break;case"gotpointercapture":case"lostpointercapture":case"pointercancel":case"pointerdown":case"pointermove":case"pointerout":case"pointerover":case"pointerup":m=v1}var d=(t&4)!==0,g=!d&&e==="scroll",v=d?f!==null?f+"Capture":null:f;d=[];for(var y=u,x;y!==null;){x=y;var P=x.stateNode;if(x.tag===5&&P!==null&&(x=P,v!==null&&(P=dl(y,v),P!=null&&d.push(wl(y,P,x)))),g)break;y=y.return}0<d.length&&(f=new m(f,h,null,n,p),c.push({event:f,listeners:d}))}}if(!(t&7)){e:{if(f=e==="mouseover"||e==="pointerover",m=e==="mouseout"||e==="pointerout",f&&n!==ah&&(h=n.relatedTarget||n.fromElement)&&(qa(h)||h[Mr]))break e;if((m||f)&&(f=p.window===p?p:(f=p.ownerDocument)?f.defaultView||f.parentWindow:window,m?(h=n.relatedTarget||n.toElement,m=u,h=h?qa(h):null,h!==null&&(g=ko(h),h!==g||h.tag!==5&&h.tag!==6)&&(h=null)):(m=null,h=u),m!==h)){if(d=m1,P="onMouseLeave",v="onMouseEnter",y="mouse",(e==="pointerout"||e==="pointerover")&&(d=v1,P="onPointerLeave",v="onPointerEnter",y="pointer"),g=m==null?f:Uo(m),x=h==null?f:Uo(h),f=new d(P,y+"leave",m,n,p),f.target=g,f.relatedTarget=x,P=null,qa(p)===u&&(d=new d(v,y+"enter",h,n,p),d.target=x,d.relatedTarget=g,P=d),g=P,m&&h)t:{for(d=m,v=h,y=0,x=d;x;x=jo(x))y++;for(x=0,P=v;P;P=jo(P))x++;for(;0<y-x;)d=jo(d),y--;for(;0<x-y;)v=jo(v),x--;for(;y--;){if(d===v||v!==null&&d===v.alternate)break t;d=jo(d),v=jo(v)}d=null}else d=null;m!==null&&_1(c,f,m,d,!1),h!==null&&g!==null&&_1(c,g,h,d,!0)}}e:{if(f=u?Uo(u):window,m=f.nodeName&&f.nodeName.toLowerCase(),m==="select"||m==="input"&&f.type==="file")var w=Lk;else if(x1(f))if(L5)w=Hk;else{w=Bk;var S=Fk}else(m=f.nodeName)&&m.toLowerCase()==="input"&&(f.type==="checkbox"||f.type==="radio")&&(w=zk);if(w&&(w=w(e,u))){D5(c,w,n,p);break e}S&&S(e,f,u),e==="focusout"&&(S=f._wrapperState)&&S.controlled&&f.type==="number"&&Jm(f,"number",f.value)}switch(S=u?Uo(u):window,e){case"focusin":(x1(S)||S.contentEditable==="true")&&(Ho=S,ph=u,tl=null);break;case"focusout":tl=ph=Ho=null;break;case"mousedown":fh=!0;break;case"contextmenu":case"mouseup":case"dragend":fh=!1,O1(c,n,p);break;case"selectionchange":if(Wk)break;case"keydown":case"keyup":O1(c,n,p)}var O;if(Gv)e:{switch(e){case"compositionstart":var C="onCompositionStart";break e;case"compositionend":C="onCompositionEnd";break e;case"compositionupdate":C="onCompositionUpdate";break e}C=void 0}else zo?R5(e,n)&&(C="onCompositionEnd"):e==="keydown"&&n.keyCode===229&&(C="onCompositionStart");C&&(M5&&n.locale!=="ko"&&(zo||C!=="onCompositionStart"?C==="onCompositionEnd"&&zo&&(O=$5()):(ua=p,Bv="value"in ua?ua.value:ua.textContent,zo=!0)),S=Uc(u,C),0<S.length&&(C=new h1(C,e,null,n,p),c.push({event:C,listeners:S}),O?C.data=O:(O=I5(n),O!==null&&(C.data=O)))),(O=$k?Mk(e,n):Rk(e,n))&&(u=Uc(u,"onBeforeInput"),0<u.length&&(p=new h1("onBeforeInput","beforeinput",null,n,p),c.push({event:p,listeners:u}),p.data=O))}K5(c,t)})}function wl(e,t,n){return{instance:e,listener:t,currentTarget:n}}function Uc(e,t){for(var n=t+"Capture",r=[];e!==null;){var a=e,o=a.stateNode;a.tag===5&&o!==null&&(a=o,o=dl(e,n),o!=null&&r.unshift(wl(e,o,a)),o=dl(e,t),o!=null&&r.push(wl(e,o,a))),e=e.return}return r}function jo(e){if(e===null)return null;do e=e.return;while(e&&e.tag!==5);return e||null}function _1(e,t,n,r,a){for(var o=t._reactName,i=[];n!==null&&n!==r;){var s=n,l=s.alternate,u=s.stateNode;if(l!==null&&l===r)break;s.tag===5&&u!==null&&(s=u,a?(l=dl(n,o),l!=null&&i.unshift(wl(n,l,s))):a||(l=dl(n,o),l!=null&&i.push(wl(n,l,s)))),n=n.return}i.length!==0&&e.push({event:t,listeners:i})}var Xk=/\r\n?/g,Yk=/\u0000|\uFFFD/g;function A1(e){return(typeof e=="string"?e:""+e).replace(Xk,`
`).replace(Yk,"")}function Wu(e,t,n){if(t=A1(t),A1(e)!==t&&n)throw Error(U(425))}function Wc(){}var dh=null,mh=null;function hh(e,t){return e==="textarea"||e==="noscript"||typeof t.children=="string"||typeof t.children=="number"||typeof t.dangerouslySetInnerHTML=="object"&&t.dangerouslySetInnerHTML!==null&&t.dangerouslySetInnerHTML.__html!=null}var vh=typeof setTimeout=="function"?setTimeout:void 0,Qk=typeof clearTimeout=="function"?clearTimeout:void 0,E1=typeof Promise=="function"?Promise:void 0,Zk=typeof queueMicrotask=="function"?queueMicrotask:typeof E1<"u"?function(e){return E1.resolve(null).then(e).catch(Jk)}:vh;function Jk(e){setTimeout(function(){throw e})}function Yd(e,t){var n=t,r=0;do{var a=n.nextSibling;if(e.removeChild(n),a&&a.nodeType===8)if(n=a.data,n==="/$"){if(r===0){e.removeChild(a),vl(t);return}r--}else n!=="$"&&n!=="$?"&&n!=="$!"||r++;n=a}while(n);vl(t)}function va(e){for(;e!=null;e=e.nextSibling){var t=e.nodeType;if(t===1||t===3)break;if(t===8){if(t=e.data,t==="$"||t==="$!"||t==="$?")break;if(t==="/$")return null}}return e}function T1(e){e=e.previousSibling;for(var t=0;e;){if(e.nodeType===8){var n=e.data;if(n==="$"||n==="$!"||n==="$?"){if(t===0)return e;t--}else n==="/$"&&t++}e=e.previousSibling}return null}var as=Math.random().toString(36).slice(2),er="__reactFiber$"+as,bl="__reactProps$"+as,Mr="__reactContainer$"+as,yh="__reactEvents$"+as,eC="__reactListeners$"+as,tC="__reactHandles$"+as;function qa(e){var t=e[er];if(t)return t;for(var n=e.parentNode;n;){if(t=n[Mr]||n[er]){if(n=t.alternate,t.child!==null||n!==null&&n.child!==null)for(e=T1(e);e!==null;){if(n=e[er])return n;e=T1(e)}return t}e=n,n=e.parentNode}return null}function wu(e){return e=e[er]||e[Mr],!e||e.tag!==5&&e.tag!==6&&e.tag!==13&&e.tag!==3?null:e}function Uo(e){if(e.tag===5||e.tag===6)return e.stateNode;throw Error(U(33))}function Pf(e){return e[bl]||null}var gh=[],Wo=-1;function Aa(e){return{current:e}}function Fe(e){0>Wo||(e.current=gh[Wo],gh[Wo]=null,Wo--)}function je(e,t){Wo++,gh[Wo]=e.current,e.current=t}var Sa={},At=Aa(Sa),Gt=Aa(!1),co=Sa;function Oi(e,t){var n=e.type.contextTypes;if(!n)return Sa;var r=e.stateNode;if(r&&r.__reactInternalMemoizedUnmaskedChildContext===t)return r.__reactInternalMemoizedMaskedChildContext;var a={},o;for(o in n)a[o]=t[o];return r&&(e=e.stateNode,e.__reactInternalMemoizedUnmaskedChildContext=t,e.__reactInternalMemoizedMaskedChildContext=a),a}function Ut(e){return e=e.childContextTypes,e!=null}function qc(){Fe(Gt),Fe(At)}function j1(e,t,n){if(At.current!==Sa)throw Error(U(168));je(At,t),je(Gt,n)}function Y5(e,t,n){var r=e.stateNode;if(t=t.childContextTypes,typeof r.getChildContext!="function")return n;r=r.getChildContext();for(var a in r)if(!(a in t))throw Error(U(108,FO(e)||"Unknown",a));return Ue({},n,r)}function Vc(e){return e=(e=e.stateNode)&&e.__reactInternalMemoizedMergedChildContext||Sa,co=At.current,je(At,e),je(Gt,Gt.current),!0}function N1(e,t,n){var r=e.stateNode;if(!r)throw Error(U(169));n?(e=Y5(e,t,co),r.__reactInternalMemoizedMergedChildContext=e,Fe(Gt),Fe(At),je(At,e)):Fe(Gt),je(Gt,n)}var br=null,Sf=!1,Qd=!1;function Q5(e){br===null?br=[e]:br.push(e)}function nC(e){Sf=!0,Q5(e)}function Ea(){if(!Qd&&br!==null){Qd=!0;var e=0,t=Oe;try{var n=br;for(Oe=1;e<n.length;e++){var r=n[e];do r=r(!0);while(r!==null)}br=null,Sf=!1}catch(a){throw br!==null&&(br=br.slice(e+1)),P5(Iv,Ea),a}finally{Oe=t,Qd=!1}}return null}var qo=[],Vo=0,Kc=null,Xc=0,pn=[],fn=0,po=null,Sr=1,Or="";function za(e,t){qo[Vo++]=Xc,qo[Vo++]=Kc,Kc=e,Xc=t}function Z5(e,t,n){pn[fn++]=Sr,pn[fn++]=Or,pn[fn++]=po,po=e;var r=Sr;e=Or;var a=32-zn(r)-1;r&=~(1<<a),n+=1;var o=32-zn(t)+a;if(30<o){var i=a-a%5;o=(r&(1<<i)-1).toString(32),r>>=i,a-=i,Sr=1<<32-zn(t)+a|n<<a|r,Or=o+e}else Sr=1<<o|n<<a|r,Or=e}function Wv(e){e.return!==null&&(za(e,1),Z5(e,1,0))}function qv(e){for(;e===Kc;)Kc=qo[--Vo],qo[Vo]=null,Xc=qo[--Vo],qo[Vo]=null;for(;e===po;)po=pn[--fn],pn[fn]=null,Or=pn[--fn],pn[fn]=null,Sr=pn[--fn],pn[fn]=null}var tn=null,Jt=null,Be=!1,Mn=null;function J5(e,t){var n=dn(5,null,null,0);n.elementType="DELETED",n.stateNode=t,n.return=e,t=e.deletions,t===null?(e.deletions=[n],e.flags|=16):t.push(n)}function $1(e,t){switch(e.tag){case 5:var n=e.type;return t=t.nodeType!==1||n.toLowerCase()!==t.nodeName.toLowerCase()?null:t,t!==null?(e.stateNode=t,tn=e,Jt=va(t.firstChild),!0):!1;case 6:return t=e.pendingProps===""||t.nodeType!==3?null:t,t!==null?(e.stateNode=t,tn=e,Jt=null,!0):!1;case 13:return t=t.nodeType!==8?null:t,t!==null?(n=po!==null?{id:Sr,overflow:Or}:null,e.memoizedState={dehydrated:t,treeContext:n,retryLane:1073741824},n=dn(18,null,null,0),n.stateNode=t,n.return=e,e.child=n,tn=e,Jt=null,!0):!1;default:return!1}}function xh(e){return(e.mode&1)!==0&&(e.flags&128)===0}function wh(e){if(Be){var t=Jt;if(t){var n=t;if(!$1(e,t)){if(xh(e))throw Error(U(418));t=va(n.nextSibling);var r=tn;t&&$1(e,t)?J5(r,n):(e.flags=e.flags&-4097|2,Be=!1,tn=e)}}else{if(xh(e))throw Error(U(418));e.flags=e.flags&-4097|2,Be=!1,tn=e}}}function M1(e){for(e=e.return;e!==null&&e.tag!==5&&e.tag!==3&&e.tag!==13;)e=e.return;tn=e}function qu(e){if(e!==tn)return!1;if(!Be)return M1(e),Be=!0,!1;var t;if((t=e.tag!==3)&&!(t=e.tag!==5)&&(t=e.type,t=t!=="head"&&t!=="body"&&!hh(e.type,e.memoizedProps)),t&&(t=Jt)){if(xh(e))throw e9(),Error(U(418));for(;t;)J5(e,t),t=va(t.nextSibling)}if(M1(e),e.tag===13){if(e=e.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(U(317));e:{for(e=e.nextSibling,t=0;e;){if(e.nodeType===8){var n=e.data;if(n==="/$"){if(t===0){Jt=va(e.nextSibling);break e}t--}else n!=="$"&&n!=="$!"&&n!=="$?"||t++}e=e.nextSibling}Jt=null}}else Jt=tn?va(e.stateNode.nextSibling):null;return!0}function e9(){for(var e=Jt;e;)e=va(e.nextSibling)}function ki(){Jt=tn=null,Be=!1}function Vv(e){Mn===null?Mn=[e]:Mn.push(e)}var rC=Hr.ReactCurrentBatchConfig;function Ts(e,t,n){if(e=n.ref,e!==null&&typeof e!="function"&&typeof e!="object"){if(n._owner){if(n=n._owner,n){if(n.tag!==1)throw Error(U(309));var r=n.stateNode}if(!r)throw Error(U(147,e));var a=r,o=""+e;return t!==null&&t.ref!==null&&typeof t.ref=="function"&&t.ref._stringRef===o?t.ref:(t=function(i){var s=a.refs;i===null?delete s[o]:s[o]=i},t._stringRef=o,t)}if(typeof e!="string")throw Error(U(284));if(!n._owner)throw Error(U(290,e))}return e}function Vu(e,t){throw e=Object.prototype.toString.call(t),Error(U(31,e==="[object Object]"?"object with keys {"+Object.keys(t).join(", ")+"}":e))}function R1(e){var t=e._init;return t(e._payload)}function t9(e){function t(v,y){if(e){var x=v.deletions;x===null?(v.deletions=[y],v.flags|=16):x.push(y)}}function n(v,y){if(!e)return null;for(;y!==null;)t(v,y),y=y.sibling;return null}function r(v,y){for(v=new Map;y!==null;)y.key!==null?v.set(y.key,y):v.set(y.index,y),y=y.sibling;return v}function a(v,y){return v=wa(v,y),v.index=0,v.sibling=null,v}function o(v,y,x){return v.index=x,e?(x=v.alternate,x!==null?(x=x.index,x<y?(v.flags|=2,y):x):(v.flags|=2,y)):(v.flags|=1048576,y)}function i(v){return e&&v.alternate===null&&(v.flags|=2),v}function s(v,y,x,P){return y===null||y.tag!==6?(y=am(x,v.mode,P),y.return=v,y):(y=a(y,x),y.return=v,y)}function l(v,y,x,P){var w=x.type;return w===Bo?p(v,y,x.props.children,P,x.key):y!==null&&(y.elementType===w||typeof w=="object"&&w!==null&&w.$$typeof===Zr&&R1(w)===y.type)?(P=a(y,x.props),P.ref=Ts(v,y,x),P.return=v,P):(P=Tc(x.type,x.key,x.props,null,v.mode,P),P.ref=Ts(v,y,x),P.return=v,P)}function u(v,y,x,P){return y===null||y.tag!==4||y.stateNode.containerInfo!==x.containerInfo||y.stateNode.implementation!==x.implementation?(y=om(x,v.mode,P),y.return=v,y):(y=a(y,x.children||[]),y.return=v,y)}function p(v,y,x,P,w){return y===null||y.tag!==7?(y=so(x,v.mode,P,w),y.return=v,y):(y=a(y,x),y.return=v,y)}function c(v,y,x){if(typeof y=="string"&&y!==""||typeof y=="number")return y=am(""+y,v.mode,x),y.return=v,y;if(typeof y=="object"&&y!==null){switch(y.$$typeof){case Iu:return x=Tc(y.type,y.key,y.props,null,v.mode,x),x.ref=Ts(v,null,y),x.return=v,x;case Fo:return y=om(y,v.mode,x),y.return=v,y;case Zr:var P=y._init;return c(v,P(y._payload),x)}if(qs(y)||ks(y))return y=so(y,v.mode,x,null),y.return=v,y;Vu(v,y)}return null}function f(v,y,x,P){var w=y!==null?y.key:null;if(typeof x=="string"&&x!==""||typeof x=="number")return w!==null?null:s(v,y,""+x,P);if(typeof x=="object"&&x!==null){switch(x.$$typeof){case Iu:return x.key===w?l(v,y,x,P):null;case Fo:return x.key===w?u(v,y,x,P):null;case Zr:return w=x._init,f(v,y,w(x._payload),P)}if(qs(x)||ks(x))return w!==null?null:p(v,y,x,P,null);Vu(v,x)}return null}function m(v,y,x,P,w){if(typeof P=="string"&&P!==""||typeof P=="number")return v=v.get(x)||null,s(y,v,""+P,w);if(typeof P=="object"&&P!==null){switch(P.$$typeof){case Iu:return v=v.get(P.key===null?x:P.key)||null,l(y,v,P,w);case Fo:return v=v.get(P.key===null?x:P.key)||null,u(y,v,P,w);case Zr:var S=P._init;return m(v,y,x,S(P._payload),w)}if(qs(P)||ks(P))return v=v.get(x)||null,p(y,v,P,w,null);Vu(y,P)}return null}function h(v,y,x,P){for(var w=null,S=null,O=y,C=y=0,_=null;O!==null&&C<x.length;C++){O.index>C?(_=O,O=null):_=O.sibling;var T=f(v,O,x[C],P);if(T===null){O===null&&(O=_);break}e&&O&&T.alternate===null&&t(v,O),y=o(T,y,C),S===null?w=T:S.sibling=T,S=T,O=_}if(C===x.length)return n(v,O),Be&&za(v,C),w;if(O===null){for(;C<x.length;C++)O=c(v,x[C],P),O!==null&&(y=o(O,y,C),S===null?w=O:S.sibling=O,S=O);return Be&&za(v,C),w}for(O=r(v,O);C<x.length;C++)_=m(O,v,C,x[C],P),_!==null&&(e&&_.alternate!==null&&O.delete(_.key===null?C:_.key),y=o(_,y,C),S===null?w=_:S.sibling=_,S=_);return e&&O.forEach(function(A){return t(v,A)}),Be&&za(v,C),w}function d(v,y,x,P){var w=ks(x);if(typeof w!="function")throw Error(U(150));if(x=w.call(x),x==null)throw Error(U(151));for(var S=w=null,O=y,C=y=0,_=null,T=x.next();O!==null&&!T.done;C++,T=x.next()){O.index>C?(_=O,O=null):_=O.sibling;var A=f(v,O,T.value,P);if(A===null){O===null&&(O=_);break}e&&O&&A.alternate===null&&t(v,O),y=o(A,y,C),S===null?w=A:S.sibling=A,S=A,O=_}if(T.done)return n(v,O),Be&&za(v,C),w;if(O===null){for(;!T.done;C++,T=x.next())T=c(v,T.value,P),T!==null&&(y=o(T,y,C),S===null?w=T:S.sibling=T,S=T);return Be&&za(v,C),w}for(O=r(v,O);!T.done;C++,T=x.next())T=m(O,v,C,T.value,P),T!==null&&(e&&T.alternate!==null&&O.delete(T.key===null?C:T.key),y=o(T,y,C),S===null?w=T:S.sibling=T,S=T);return e&&O.forEach(function(j){return t(v,j)}),Be&&za(v,C),w}function g(v,y,x,P){if(typeof x=="object"&&x!==null&&x.type===Bo&&x.key===null&&(x=x.props.children),typeof x=="object"&&x!==null){switch(x.$$typeof){case Iu:e:{for(var w=x.key,S=y;S!==null;){if(S.key===w){if(w=x.type,w===Bo){if(S.tag===7){n(v,S.sibling),y=a(S,x.props.children),y.return=v,v=y;break e}}else if(S.elementType===w||typeof w=="object"&&w!==null&&w.$$typeof===Zr&&R1(w)===S.type){n(v,S.sibling),y=a(S,x.props),y.ref=Ts(v,S,x),y.return=v,v=y;break e}n(v,S);break}else t(v,S);S=S.sibling}x.type===Bo?(y=so(x.props.children,v.mode,P,x.key),y.return=v,v=y):(P=Tc(x.type,x.key,x.props,null,v.mode,P),P.ref=Ts(v,y,x),P.return=v,v=P)}return i(v);case Fo:e:{for(S=x.key;y!==null;){if(y.key===S)if(y.tag===4&&y.stateNode.containerInfo===x.containerInfo&&y.stateNode.implementation===x.implementation){n(v,y.sibling),y=a(y,x.children||[]),y.return=v,v=y;break e}else{n(v,y);break}else t(v,y);y=y.sibling}y=om(x,v.mode,P),y.return=v,v=y}return i(v);case Zr:return S=x._init,g(v,y,S(x._payload),P)}if(qs(x))return h(v,y,x,P);if(ks(x))return d(v,y,x,P);Vu(v,x)}return typeof x=="string"&&x!==""||typeof x=="number"?(x=""+x,y!==null&&y.tag===6?(n(v,y.sibling),y=a(y,x),y.return=v,v=y):(n(v,y),y=am(x,v.mode,P),y.return=v,v=y),i(v)):n(v,y)}return g}var Ci=t9(!0),n9=t9(!1),Yc=Aa(null),Qc=null,Ko=null,Kv=null;function Xv(){Kv=Ko=Qc=null}function Yv(e){var t=Yc.current;Fe(Yc),e._currentValue=t}function bh(e,t,n){for(;e!==null;){var r=e.alternate;if((e.childLanes&t)!==t?(e.childLanes|=t,r!==null&&(r.childLanes|=t)):r!==null&&(r.childLanes&t)!==t&&(r.childLanes|=t),e===n)break;e=e.return}}function ai(e,t){Qc=e,Kv=Ko=null,e=e.dependencies,e!==null&&e.firstContext!==null&&(e.lanes&t&&(zt=!0),e.firstContext=null)}function gn(e){var t=e._currentValue;if(Kv!==e)if(e={context:e,memoizedValue:t,next:null},Ko===null){if(Qc===null)throw Error(U(308));Ko=e,Qc.dependencies={lanes:0,firstContext:e}}else Ko=Ko.next=e;return t}var Va=null;function Qv(e){Va===null?Va=[e]:Va.push(e)}function r9(e,t,n,r){var a=t.interleaved;return a===null?(n.next=n,Qv(t)):(n.next=a.next,a.next=n),t.interleaved=n,Rr(e,r)}function Rr(e,t){e.lanes|=t;var n=e.alternate;for(n!==null&&(n.lanes|=t),n=e,e=e.return;e!==null;)e.childLanes|=t,n=e.alternate,n!==null&&(n.childLanes|=t),n=e,e=e.return;return n.tag===3?n.stateNode:null}var Jr=!1;function Zv(e){e.updateQueue={baseState:e.memoizedState,firstBaseUpdate:null,lastBaseUpdate:null,shared:{pending:null,interleaved:null,lanes:0},effects:null}}function a9(e,t){e=e.updateQueue,t.updateQueue===e&&(t.updateQueue={baseState:e.baseState,firstBaseUpdate:e.firstBaseUpdate,lastBaseUpdate:e.lastBaseUpdate,shared:e.shared,effects:e.effects})}function Ar(e,t){return{eventTime:e,lane:t,tag:0,payload:null,callback:null,next:null}}function ya(e,t,n){var r=e.updateQueue;if(r===null)return null;if(r=r.shared,de&2){var a=r.pending;return a===null?t.next=t:(t.next=a.next,a.next=t),r.pending=t,Rr(e,n)}return a=r.interleaved,a===null?(t.next=t,Qv(r)):(t.next=a.next,a.next=t),r.interleaved=t,Rr(e,n)}function Oc(e,t,n){if(t=t.updateQueue,t!==null&&(t=t.shared,(n&4194240)!==0)){var r=t.lanes;r&=e.pendingLanes,n|=r,t.lanes=n,Dv(e,n)}}function I1(e,t){var n=e.updateQueue,r=e.alternate;if(r!==null&&(r=r.updateQueue,n===r)){var a=null,o=null;if(n=n.firstBaseUpdate,n!==null){do{var i={eventTime:n.eventTime,lane:n.lane,tag:n.tag,payload:n.payload,callback:n.callback,next:null};o===null?a=o=i:o=o.next=i,n=n.next}while(n!==null);o===null?a=o=t:o=o.next=t}else a=o=t;n={baseState:r.baseState,firstBaseUpdate:a,lastBaseUpdate:o,shared:r.shared,effects:r.effects},e.updateQueue=n;return}e=n.lastBaseUpdate,e===null?n.firstBaseUpdate=t:e.next=t,n.lastBaseUpdate=t}function Zc(e,t,n,r){var a=e.updateQueue;Jr=!1;var o=a.firstBaseUpdate,i=a.lastBaseUpdate,s=a.shared.pending;if(s!==null){a.shared.pending=null;var l=s,u=l.next;l.next=null,i===null?o=u:i.next=u,i=l;var p=e.alternate;p!==null&&(p=p.updateQueue,s=p.lastBaseUpdate,s!==i&&(s===null?p.firstBaseUpdate=u:s.next=u,p.lastBaseUpdate=l))}if(o!==null){var c=a.baseState;i=0,p=u=l=null,s=o;do{var f=s.lane,m=s.eventTime;if((r&f)===f){p!==null&&(p=p.next={eventTime:m,lane:0,tag:s.tag,payload:s.payload,callback:s.callback,next:null});e:{var h=e,d=s;switch(f=t,m=n,d.tag){case 1:if(h=d.payload,typeof h=="function"){c=h.call(m,c,f);break e}c=h;break e;case 3:h.flags=h.flags&-65537|128;case 0:if(h=d.payload,f=typeof h=="function"?h.call(m,c,f):h,f==null)break e;c=Ue({},c,f);break e;case 2:Jr=!0}}s.callback!==null&&s.lane!==0&&(e.flags|=64,f=a.effects,f===null?a.effects=[s]:f.push(s))}else m={eventTime:m,lane:f,tag:s.tag,payload:s.payload,callback:s.callback,next:null},p===null?(u=p=m,l=c):p=p.next=m,i|=f;if(s=s.next,s===null){if(s=a.shared.pending,s===null)break;f=s,s=f.next,f.next=null,a.lastBaseUpdate=f,a.shared.pending=null}}while(!0);if(p===null&&(l=c),a.baseState=l,a.firstBaseUpdate=u,a.lastBaseUpdate=p,t=a.shared.interleaved,t!==null){a=t;do i|=a.lane,a=a.next;while(a!==t)}else o===null&&(a.shared.lanes=0);mo|=i,e.lanes=i,e.memoizedState=c}}function D1(e,t,n){if(e=t.effects,t.effects=null,e!==null)for(t=0;t<e.length;t++){var r=e[t],a=r.callback;if(a!==null){if(r.callback=null,r=n,typeof a!="function")throw Error(U(191,a));a.call(r)}}}var bu={},lr=Aa(bu),Pl=Aa(bu),Sl=Aa(bu);function Ka(e){if(e===bu)throw Error(U(174));return e}function Jv(e,t){switch(je(Sl,t),je(Pl,e),je(lr,bu),e=t.nodeType,e){case 9:case 11:t=(t=t.documentElement)?t.namespaceURI:th(null,"");break;default:e=e===8?t.parentNode:t,t=e.namespaceURI||null,e=e.tagName,t=th(t,e)}Fe(lr),je(lr,t)}function _i(){Fe(lr),Fe(Pl),Fe(Sl)}function o9(e){Ka(Sl.current);var t=Ka(lr.current),n=th(t,e.type);t!==n&&(je(Pl,e),je(lr,n))}function ey(e){Pl.current===e&&(Fe(lr),Fe(Pl))}var He=Aa(0);function Jc(e){for(var t=e;t!==null;){if(t.tag===13){var n=t.memoizedState;if(n!==null&&(n=n.dehydrated,n===null||n.data==="$?"||n.data==="$!"))return t}else if(t.tag===19&&t.memoizedProps.revealOrder!==void 0){if(t.flags&128)return t}else if(t.child!==null){t.child.return=t,t=t.child;continue}if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return null;t=t.return}t.sibling.return=t.return,t=t.sibling}return null}var Zd=[];function ty(){for(var e=0;e<Zd.length;e++)Zd[e]._workInProgressVersionPrimary=null;Zd.length=0}var kc=Hr.ReactCurrentDispatcher,Jd=Hr.ReactCurrentBatchConfig,fo=0,Ge=null,rt=null,lt=null,ep=!1,nl=!1,Ol=0,aC=0;function xt(){throw Error(U(321))}function ny(e,t){if(t===null)return!1;for(var n=0;n<t.length&&n<e.length;n++)if(!Gn(e[n],t[n]))return!1;return!0}function ry(e,t,n,r,a,o){if(fo=o,Ge=t,t.memoizedState=null,t.updateQueue=null,t.lanes=0,kc.current=e===null||e.memoizedState===null?lC:uC,e=n(r,a),nl){o=0;do{if(nl=!1,Ol=0,25<=o)throw Error(U(301));o+=1,lt=rt=null,t.updateQueue=null,kc.current=cC,e=n(r,a)}while(nl)}if(kc.current=tp,t=rt!==null&&rt.next!==null,fo=0,lt=rt=Ge=null,ep=!1,t)throw Error(U(300));return e}function ay(){var e=Ol!==0;return Ol=0,e}function Xn(){var e={memoizedState:null,baseState:null,baseQueue:null,queue:null,next:null};return lt===null?Ge.memoizedState=lt=e:lt=lt.next=e,lt}function xn(){if(rt===null){var e=Ge.alternate;e=e!==null?e.memoizedState:null}else e=rt.next;var t=lt===null?Ge.memoizedState:lt.next;if(t!==null)lt=t,rt=e;else{if(e===null)throw Error(U(310));rt=e,e={memoizedState:rt.memoizedState,baseState:rt.baseState,baseQueue:rt.baseQueue,queue:rt.queue,next:null},lt===null?Ge.memoizedState=lt=e:lt=lt.next=e}return lt}function kl(e,t){return typeof t=="function"?t(e):t}function em(e){var t=xn(),n=t.queue;if(n===null)throw Error(U(311));n.lastRenderedReducer=e;var r=rt,a=r.baseQueue,o=n.pending;if(o!==null){if(a!==null){var i=a.next;a.next=o.next,o.next=i}r.baseQueue=a=o,n.pending=null}if(a!==null){o=a.next,r=r.baseState;var s=i=null,l=null,u=o;do{var p=u.lane;if((fo&p)===p)l!==null&&(l=l.next={lane:0,action:u.action,hasEagerState:u.hasEagerState,eagerState:u.eagerState,next:null}),r=u.hasEagerState?u.eagerState:e(r,u.action);else{var c={lane:p,action:u.action,hasEagerState:u.hasEagerState,eagerState:u.eagerState,next:null};l===null?(s=l=c,i=r):l=l.next=c,Ge.lanes|=p,mo|=p}u=u.next}while(u!==null&&u!==o);l===null?i=r:l.next=s,Gn(r,t.memoizedState)||(zt=!0),t.memoizedState=r,t.baseState=i,t.baseQueue=l,n.lastRenderedState=r}if(e=n.interleaved,e!==null){a=e;do o=a.lane,Ge.lanes|=o,mo|=o,a=a.next;while(a!==e)}else a===null&&(n.lanes=0);return[t.memoizedState,n.dispatch]}function tm(e){var t=xn(),n=t.queue;if(n===null)throw Error(U(311));n.lastRenderedReducer=e;var r=n.dispatch,a=n.pending,o=t.memoizedState;if(a!==null){n.pending=null;var i=a=a.next;do o=e(o,i.action),i=i.next;while(i!==a);Gn(o,t.memoizedState)||(zt=!0),t.memoizedState=o,t.baseQueue===null&&(t.baseState=o),n.lastRenderedState=o}return[o,r]}function i9(){}function s9(e,t){var n=Ge,r=xn(),a=t(),o=!Gn(r.memoizedState,a);if(o&&(r.memoizedState=a,zt=!0),r=r.queue,oy(c9.bind(null,n,r,e),[e]),r.getSnapshot!==t||o||lt!==null&&lt.memoizedState.tag&1){if(n.flags|=2048,Cl(9,u9.bind(null,n,r,a,t),void 0,null),ut===null)throw Error(U(349));fo&30||l9(n,t,a)}return a}function l9(e,t,n){e.flags|=16384,e={getSnapshot:t,value:n},t=Ge.updateQueue,t===null?(t={lastEffect:null,stores:null},Ge.updateQueue=t,t.stores=[e]):(n=t.stores,n===null?t.stores=[e]:n.push(e))}function u9(e,t,n,r){t.value=n,t.getSnapshot=r,p9(t)&&f9(e)}function c9(e,t,n){return n(function(){p9(t)&&f9(e)})}function p9(e){var t=e.getSnapshot;e=e.value;try{var n=t();return!Gn(e,n)}catch{return!0}}function f9(e){var t=Rr(e,1);t!==null&&Hn(t,e,1,-1)}function L1(e){var t=Xn();return typeof e=="function"&&(e=e()),t.memoizedState=t.baseState=e,e={pending:null,interleaved:null,lanes:0,dispatch:null,lastRenderedReducer:kl,lastRenderedState:e},t.queue=e,e=e.dispatch=sC.bind(null,Ge,e),[t.memoizedState,e]}function Cl(e,t,n,r){return e={tag:e,create:t,destroy:n,deps:r,next:null},t=Ge.updateQueue,t===null?(t={lastEffect:null,stores:null},Ge.updateQueue=t,t.lastEffect=e.next=e):(n=t.lastEffect,n===null?t.lastEffect=e.next=e:(r=n.next,n.next=e,e.next=r,t.lastEffect=e)),e}function d9(){return xn().memoizedState}function Cc(e,t,n,r){var a=Xn();Ge.flags|=e,a.memoizedState=Cl(1|t,n,void 0,r===void 0?null:r)}function Of(e,t,n,r){var a=xn();r=r===void 0?null:r;var o=void 0;if(rt!==null){var i=rt.memoizedState;if(o=i.destroy,r!==null&&ny(r,i.deps)){a.memoizedState=Cl(t,n,o,r);return}}Ge.flags|=e,a.memoizedState=Cl(1|t,n,o,r)}function F1(e,t){return Cc(8390656,8,e,t)}function oy(e,t){return Of(2048,8,e,t)}function m9(e,t){return Of(4,2,e,t)}function h9(e,t){return Of(4,4,e,t)}function v9(e,t){if(typeof t=="function")return e=e(),t(e),function(){t(null)};if(t!=null)return e=e(),t.current=e,function(){t.current=null}}function y9(e,t,n){return n=n!=null?n.concat([e]):null,Of(4,4,v9.bind(null,t,e),n)}function iy(){}function g9(e,t){var n=xn();t=t===void 0?null:t;var r=n.memoizedState;return r!==null&&t!==null&&ny(t,r[1])?r[0]:(n.memoizedState=[e,t],e)}function x9(e,t){var n=xn();t=t===void 0?null:t;var r=n.memoizedState;return r!==null&&t!==null&&ny(t,r[1])?r[0]:(e=e(),n.memoizedState=[e,t],e)}function w9(e,t,n){return fo&21?(Gn(n,t)||(n=k5(),Ge.lanes|=n,mo|=n,e.baseState=!0),t):(e.baseState&&(e.baseState=!1,zt=!0),e.memoizedState=n)}function oC(e,t){var n=Oe;Oe=n!==0&&4>n?n:4,e(!0);var r=Jd.transition;Jd.transition={};try{e(!1),t()}finally{Oe=n,Jd.transition=r}}function b9(){return xn().memoizedState}function iC(e,t,n){var r=xa(e);if(n={lane:r,action:n,hasEagerState:!1,eagerState:null,next:null},P9(e))S9(t,n);else if(n=r9(e,t,n,r),n!==null){var a=Rt();Hn(n,e,r,a),O9(n,t,r)}}function sC(e,t,n){var r=xa(e),a={lane:r,action:n,hasEagerState:!1,eagerState:null,next:null};if(P9(e))S9(t,a);else{var o=e.alternate;if(e.lanes===0&&(o===null||o.lanes===0)&&(o=t.lastRenderedReducer,o!==null))try{var i=t.lastRenderedState,s=o(i,n);if(a.hasEagerState=!0,a.eagerState=s,Gn(s,i)){var l=t.interleaved;l===null?(a.next=a,Qv(t)):(a.next=l.next,l.next=a),t.interleaved=a;return}}catch{}finally{}n=r9(e,t,a,r),n!==null&&(a=Rt(),Hn(n,e,r,a),O9(n,t,r))}}function P9(e){var t=e.alternate;return e===Ge||t!==null&&t===Ge}function S9(e,t){nl=ep=!0;var n=e.pending;n===null?t.next=t:(t.next=n.next,n.next=t),e.pending=t}function O9(e,t,n){if(n&4194240){var r=t.lanes;r&=e.pendingLanes,n|=r,t.lanes=n,Dv(e,n)}}var tp={readContext:gn,useCallback:xt,useContext:xt,useEffect:xt,useImperativeHandle:xt,useInsertionEffect:xt,useLayoutEffect:xt,useMemo:xt,useReducer:xt,useRef:xt,useState:xt,useDebugValue:xt,useDeferredValue:xt,useTransition:xt,useMutableSource:xt,useSyncExternalStore:xt,useId:xt,unstable_isNewReconciler:!1},lC={readContext:gn,useCallback:function(e,t){return Xn().memoizedState=[e,t===void 0?null:t],e},useContext:gn,useEffect:F1,useImperativeHandle:function(e,t,n){return n=n!=null?n.concat([e]):null,Cc(4194308,4,v9.bind(null,t,e),n)},useLayoutEffect:function(e,t){return Cc(4194308,4,e,t)},useInsertionEffect:function(e,t){return Cc(4,2,e,t)},useMemo:function(e,t){var n=Xn();return t=t===void 0?null:t,e=e(),n.memoizedState=[e,t],e},useReducer:function(e,t,n){var r=Xn();return t=n!==void 0?n(t):t,r.memoizedState=r.baseState=t,e={pending:null,interleaved:null,lanes:0,dispatch:null,lastRenderedReducer:e,lastRenderedState:t},r.queue=e,e=e.dispatch=iC.bind(null,Ge,e),[r.memoizedState,e]},useRef:function(e){var t=Xn();return e={current:e},t.memoizedState=e},useState:L1,useDebugValue:iy,useDeferredValue:function(e){return Xn().memoizedState=e},useTransition:function(){var e=L1(!1),t=e[0];return e=oC.bind(null,e[1]),Xn().memoizedState=e,[t,e]},useMutableSource:function(){},useSyncExternalStore:function(e,t,n){var r=Ge,a=Xn();if(Be){if(n===void 0)throw Error(U(407));n=n()}else{if(n=t(),ut===null)throw Error(U(349));fo&30||l9(r,t,n)}a.memoizedState=n;var o={value:n,getSnapshot:t};return a.queue=o,F1(c9.bind(null,r,o,e),[e]),r.flags|=2048,Cl(9,u9.bind(null,r,o,n,t),void 0,null),n},useId:function(){var e=Xn(),t=ut.identifierPrefix;if(Be){var n=Or,r=Sr;n=(r&~(1<<32-zn(r)-1)).toString(32)+n,t=":"+t+"R"+n,n=Ol++,0<n&&(t+="H"+n.toString(32)),t+=":"}else n=aC++,t=":"+t+"r"+n.toString(32)+":";return e.memoizedState=t},unstable_isNewReconciler:!1},uC={readContext:gn,useCallback:g9,useContext:gn,useEffect:oy,useImperativeHandle:y9,useInsertionEffect:m9,useLayoutEffect:h9,useMemo:x9,useReducer:em,useRef:d9,useState:function(){return em(kl)},useDebugValue:iy,useDeferredValue:function(e){var t=xn();return w9(t,rt.memoizedState,e)},useTransition:function(){var e=em(kl)[0],t=xn().memoizedState;return[e,t]},useMutableSource:i9,useSyncExternalStore:s9,useId:b9,unstable_isNewReconciler:!1},cC={readContext:gn,useCallback:g9,useContext:gn,useEffect:oy,useImperativeHandle:y9,useInsertionEffect:m9,useLayoutEffect:h9,useMemo:x9,useReducer:tm,useRef:d9,useState:function(){return tm(kl)},useDebugValue:iy,useDeferredValue:function(e){var t=xn();return rt===null?t.memoizedState=e:w9(t,rt.memoizedState,e)},useTransition:function(){var e=tm(kl)[0],t=xn().memoizedState;return[e,t]},useMutableSource:i9,useSyncExternalStore:s9,useId:b9,unstable_isNewReconciler:!1};function _n(e,t){if(e&&e.defaultProps){t=Ue({},t),e=e.defaultProps;for(var n in e)t[n]===void 0&&(t[n]=e[n]);return t}return t}function Ph(e,t,n,r){t=e.memoizedState,n=n(r,t),n=n==null?t:Ue({},t,n),e.memoizedState=n,e.lanes===0&&(e.updateQueue.baseState=n)}var kf={isMounted:function(e){return(e=e._reactInternals)?ko(e)===e:!1},enqueueSetState:function(e,t,n){e=e._reactInternals;var r=Rt(),a=xa(e),o=Ar(r,a);o.payload=t,n!=null&&(o.callback=n),t=ya(e,o,a),t!==null&&(Hn(t,e,a,r),Oc(t,e,a))},enqueueReplaceState:function(e,t,n){e=e._reactInternals;var r=Rt(),a=xa(e),o=Ar(r,a);o.tag=1,o.payload=t,n!=null&&(o.callback=n),t=ya(e,o,a),t!==null&&(Hn(t,e,a,r),Oc(t,e,a))},enqueueForceUpdate:function(e,t){e=e._reactInternals;var n=Rt(),r=xa(e),a=Ar(n,r);a.tag=2,t!=null&&(a.callback=t),t=ya(e,a,r),t!==null&&(Hn(t,e,r,n),Oc(t,e,r))}};function B1(e,t,n,r,a,o,i){return e=e.stateNode,typeof e.shouldComponentUpdate=="function"?e.shouldComponentUpdate(r,o,i):t.prototype&&t.prototype.isPureReactComponent?!gl(n,r)||!gl(a,o):!0}function k9(e,t,n){var r=!1,a=Sa,o=t.contextType;return typeof o=="object"&&o!==null?o=gn(o):(a=Ut(t)?co:At.current,r=t.contextTypes,o=(r=r!=null)?Oi(e,a):Sa),t=new t(n,o),e.memoizedState=t.state!==null&&t.state!==void 0?t.state:null,t.updater=kf,e.stateNode=t,t._reactInternals=e,r&&(e=e.stateNode,e.__reactInternalMemoizedUnmaskedChildContext=a,e.__reactInternalMemoizedMaskedChildContext=o),t}function z1(e,t,n,r){e=t.state,typeof t.componentWillReceiveProps=="function"&&t.componentWillReceiveProps(n,r),typeof t.UNSAFE_componentWillReceiveProps=="function"&&t.UNSAFE_componentWillReceiveProps(n,r),t.state!==e&&kf.enqueueReplaceState(t,t.state,null)}function Sh(e,t,n,r){var a=e.stateNode;a.props=n,a.state=e.memoizedState,a.refs={},Zv(e);var o=t.contextType;typeof o=="object"&&o!==null?a.context=gn(o):(o=Ut(t)?co:At.current,a.context=Oi(e,o)),a.state=e.memoizedState,o=t.getDerivedStateFromProps,typeof o=="function"&&(Ph(e,t,o,n),a.state=e.memoizedState),typeof t.getDerivedStateFromProps=="function"||typeof a.getSnapshotBeforeUpdate=="function"||typeof a.UNSAFE_componentWillMount!="function"&&typeof a.componentWillMount!="function"||(t=a.state,typeof a.componentWillMount=="function"&&a.componentWillMount(),typeof a.UNSAFE_componentWillMount=="function"&&a.UNSAFE_componentWillMount(),t!==a.state&&kf.enqueueReplaceState(a,a.state,null),Zc(e,n,a,r),a.state=e.memoizedState),typeof a.componentDidMount=="function"&&(e.flags|=4194308)}function Ai(e,t){try{var n="",r=t;do n+=LO(r),r=r.return;while(r);var a=n}catch(o){a=`
Error generating stack: `+o.message+`
`+o.stack}return{value:e,source:t,stack:a,digest:null}}function nm(e,t,n){return{value:e,source:null,stack:n??null,digest:t??null}}function Oh(e,t){try{console.error(t.value)}catch(n){setTimeout(function(){throw n})}}var pC=typeof WeakMap=="function"?WeakMap:Map;function C9(e,t,n){n=Ar(-1,n),n.tag=3,n.payload={element:null};var r=t.value;return n.callback=function(){rp||(rp=!0,Mh=r),Oh(e,t)},n}function _9(e,t,n){n=Ar(-1,n),n.tag=3;var r=e.type.getDerivedStateFromError;if(typeof r=="function"){var a=t.value;n.payload=function(){return r(a)},n.callback=function(){Oh(e,t)}}var o=e.stateNode;return o!==null&&typeof o.componentDidCatch=="function"&&(n.callback=function(){Oh(e,t),typeof r!="function"&&(ga===null?ga=new Set([this]):ga.add(this));var i=t.stack;this.componentDidCatch(t.value,{componentStack:i!==null?i:""})}),n}function H1(e,t,n){var r=e.pingCache;if(r===null){r=e.pingCache=new pC;var a=new Set;r.set(t,a)}else a=r.get(t),a===void 0&&(a=new Set,r.set(t,a));a.has(n)||(a.add(n),e=kC.bind(null,e,t,n),t.then(e,e))}function G1(e){do{var t;if((t=e.tag===13)&&(t=e.memoizedState,t=t!==null?t.dehydrated!==null:!0),t)return e;e=e.return}while(e!==null);return null}function U1(e,t,n,r,a){return e.mode&1?(e.flags|=65536,e.lanes=a,e):(e===t?e.flags|=65536:(e.flags|=128,n.flags|=131072,n.flags&=-52805,n.tag===1&&(n.alternate===null?n.tag=17:(t=Ar(-1,1),t.tag=2,ya(n,t,1))),n.lanes|=1),e)}var fC=Hr.ReactCurrentOwner,zt=!1;function jt(e,t,n,r){t.child=e===null?n9(t,null,n,r):Ci(t,e.child,n,r)}function W1(e,t,n,r,a){n=n.render;var o=t.ref;return ai(t,a),r=ry(e,t,n,r,o,a),n=ay(),e!==null&&!zt?(t.updateQueue=e.updateQueue,t.flags&=-2053,e.lanes&=~a,Ir(e,t,a)):(Be&&n&&Wv(t),t.flags|=1,jt(e,t,r,a),t.child)}function q1(e,t,n,r,a){if(e===null){var o=n.type;return typeof o=="function"&&!my(o)&&o.defaultProps===void 0&&n.compare===null&&n.defaultProps===void 0?(t.tag=15,t.type=o,A9(e,t,o,r,a)):(e=Tc(n.type,null,r,t,t.mode,a),e.ref=t.ref,e.return=t,t.child=e)}if(o=e.child,!(e.lanes&a)){var i=o.memoizedProps;if(n=n.compare,n=n!==null?n:gl,n(i,r)&&e.ref===t.ref)return Ir(e,t,a)}return t.flags|=1,e=wa(o,r),e.ref=t.ref,e.return=t,t.child=e}function A9(e,t,n,r,a){if(e!==null){var o=e.memoizedProps;if(gl(o,r)&&e.ref===t.ref)if(zt=!1,t.pendingProps=r=o,(e.lanes&a)!==0)e.flags&131072&&(zt=!0);else return t.lanes=e.lanes,Ir(e,t,a)}return kh(e,t,n,r,a)}function E9(e,t,n){var r=t.pendingProps,a=r.children,o=e!==null?e.memoizedState:null;if(r.mode==="hidden")if(!(t.mode&1))t.memoizedState={baseLanes:0,cachePool:null,transitions:null},je(Yo,Xt),Xt|=n;else{if(!(n&1073741824))return e=o!==null?o.baseLanes|n:n,t.lanes=t.childLanes=1073741824,t.memoizedState={baseLanes:e,cachePool:null,transitions:null},t.updateQueue=null,je(Yo,Xt),Xt|=e,null;t.memoizedState={baseLanes:0,cachePool:null,transitions:null},r=o!==null?o.baseLanes:n,je(Yo,Xt),Xt|=r}else o!==null?(r=o.baseLanes|n,t.memoizedState=null):r=n,je(Yo,Xt),Xt|=r;return jt(e,t,a,n),t.child}function T9(e,t){var n=t.ref;(e===null&&n!==null||e!==null&&e.ref!==n)&&(t.flags|=512,t.flags|=2097152)}function kh(e,t,n,r,a){var o=Ut(n)?co:At.current;return o=Oi(t,o),ai(t,a),n=ry(e,t,n,r,o,a),r=ay(),e!==null&&!zt?(t.updateQueue=e.updateQueue,t.flags&=-2053,e.lanes&=~a,Ir(e,t,a)):(Be&&r&&Wv(t),t.flags|=1,jt(e,t,n,a),t.child)}function V1(e,t,n,r,a){if(Ut(n)){var o=!0;Vc(t)}else o=!1;if(ai(t,a),t.stateNode===null)_c(e,t),k9(t,n,r),Sh(t,n,r,a),r=!0;else if(e===null){var i=t.stateNode,s=t.memoizedProps;i.props=s;var l=i.context,u=n.contextType;typeof u=="object"&&u!==null?u=gn(u):(u=Ut(n)?co:At.current,u=Oi(t,u));var p=n.getDerivedStateFromProps,c=typeof p=="function"||typeof i.getSnapshotBeforeUpdate=="function";c||typeof i.UNSAFE_componentWillReceiveProps!="function"&&typeof i.componentWillReceiveProps!="function"||(s!==r||l!==u)&&z1(t,i,r,u),Jr=!1;var f=t.memoizedState;i.state=f,Zc(t,r,i,a),l=t.memoizedState,s!==r||f!==l||Gt.current||Jr?(typeof p=="function"&&(Ph(t,n,p,r),l=t.memoizedState),(s=Jr||B1(t,n,s,r,f,l,u))?(c||typeof i.UNSAFE_componentWillMount!="function"&&typeof i.componentWillMount!="function"||(typeof i.componentWillMount=="function"&&i.componentWillMount(),typeof i.UNSAFE_componentWillMount=="function"&&i.UNSAFE_componentWillMount()),typeof i.componentDidMount=="function"&&(t.flags|=4194308)):(typeof i.componentDidMount=="function"&&(t.flags|=4194308),t.memoizedProps=r,t.memoizedState=l),i.props=r,i.state=l,i.context=u,r=s):(typeof i.componentDidMount=="function"&&(t.flags|=4194308),r=!1)}else{i=t.stateNode,a9(e,t),s=t.memoizedProps,u=t.type===t.elementType?s:_n(t.type,s),i.props=u,c=t.pendingProps,f=i.context,l=n.contextType,typeof l=="object"&&l!==null?l=gn(l):(l=Ut(n)?co:At.current,l=Oi(t,l));var m=n.getDerivedStateFromProps;(p=typeof m=="function"||typeof i.getSnapshotBeforeUpdate=="function")||typeof i.UNSAFE_componentWillReceiveProps!="function"&&typeof i.componentWillReceiveProps!="function"||(s!==c||f!==l)&&z1(t,i,r,l),Jr=!1,f=t.memoizedState,i.state=f,Zc(t,r,i,a);var h=t.memoizedState;s!==c||f!==h||Gt.current||Jr?(typeof m=="function"&&(Ph(t,n,m,r),h=t.memoizedState),(u=Jr||B1(t,n,u,r,f,h,l)||!1)?(p||typeof i.UNSAFE_componentWillUpdate!="function"&&typeof i.componentWillUpdate!="function"||(typeof i.componentWillUpdate=="function"&&i.componentWillUpdate(r,h,l),typeof i.UNSAFE_componentWillUpdate=="function"&&i.UNSAFE_componentWillUpdate(r,h,l)),typeof i.componentDidUpdate=="function"&&(t.flags|=4),typeof i.getSnapshotBeforeUpdate=="function"&&(t.flags|=1024)):(typeof i.componentDidUpdate!="function"||s===e.memoizedProps&&f===e.memoizedState||(t.flags|=4),typeof i.getSnapshotBeforeUpdate!="function"||s===e.memoizedProps&&f===e.memoizedState||(t.flags|=1024),t.memoizedProps=r,t.memoizedState=h),i.props=r,i.state=h,i.context=l,r=u):(typeof i.componentDidUpdate!="function"||s===e.memoizedProps&&f===e.memoizedState||(t.flags|=4),typeof i.getSnapshotBeforeUpdate!="function"||s===e.memoizedProps&&f===e.memoizedState||(t.flags|=1024),r=!1)}return Ch(e,t,n,r,o,a)}function Ch(e,t,n,r,a,o){T9(e,t);var i=(t.flags&128)!==0;if(!r&&!i)return a&&N1(t,n,!1),Ir(e,t,o);r=t.stateNode,fC.current=t;var s=i&&typeof n.getDerivedStateFromError!="function"?null:r.render();return t.flags|=1,e!==null&&i?(t.child=Ci(t,e.child,null,o),t.child=Ci(t,null,s,o)):jt(e,t,s,o),t.memoizedState=r.state,a&&N1(t,n,!0),t.child}function j9(e){var t=e.stateNode;t.pendingContext?j1(e,t.pendingContext,t.pendingContext!==t.context):t.context&&j1(e,t.context,!1),Jv(e,t.containerInfo)}function K1(e,t,n,r,a){return ki(),Vv(a),t.flags|=256,jt(e,t,n,r),t.child}var _h={dehydrated:null,treeContext:null,retryLane:0};function Ah(e){return{baseLanes:e,cachePool:null,transitions:null}}function N9(e,t,n){var r=t.pendingProps,a=He.current,o=!1,i=(t.flags&128)!==0,s;if((s=i)||(s=e!==null&&e.memoizedState===null?!1:(a&2)!==0),s?(o=!0,t.flags&=-129):(e===null||e.memoizedState!==null)&&(a|=1),je(He,a&1),e===null)return wh(t),e=t.memoizedState,e!==null&&(e=e.dehydrated,e!==null)?(t.mode&1?e.data==="$!"?t.lanes=8:t.lanes=1073741824:t.lanes=1,null):(i=r.children,e=r.fallback,o?(r=t.mode,o=t.child,i={mode:"hidden",children:i},!(r&1)&&o!==null?(o.childLanes=0,o.pendingProps=i):o=Af(i,r,0,null),e=so(e,r,n,null),o.return=t,e.return=t,o.sibling=e,t.child=o,t.child.memoizedState=Ah(n),t.memoizedState=_h,e):sy(t,i));if(a=e.memoizedState,a!==null&&(s=a.dehydrated,s!==null))return dC(e,t,i,r,s,a,n);if(o){o=r.fallback,i=t.mode,a=e.child,s=a.sibling;var l={mode:"hidden",children:r.children};return!(i&1)&&t.child!==a?(r=t.child,r.childLanes=0,r.pendingProps=l,t.deletions=null):(r=wa(a,l),r.subtreeFlags=a.subtreeFlags&14680064),s!==null?o=wa(s,o):(o=so(o,i,n,null),o.flags|=2),o.return=t,r.return=t,r.sibling=o,t.child=r,r=o,o=t.child,i=e.child.memoizedState,i=i===null?Ah(n):{baseLanes:i.baseLanes|n,cachePool:null,transitions:i.transitions},o.memoizedState=i,o.childLanes=e.childLanes&~n,t.memoizedState=_h,r}return o=e.child,e=o.sibling,r=wa(o,{mode:"visible",children:r.children}),!(t.mode&1)&&(r.lanes=n),r.return=t,r.sibling=null,e!==null&&(n=t.deletions,n===null?(t.deletions=[e],t.flags|=16):n.push(e)),t.child=r,t.memoizedState=null,r}function sy(e,t){return t=Af({mode:"visible",children:t},e.mode,0,null),t.return=e,e.child=t}function Ku(e,t,n,r){return r!==null&&Vv(r),Ci(t,e.child,null,n),e=sy(t,t.pendingProps.children),e.flags|=2,t.memoizedState=null,e}function dC(e,t,n,r,a,o,i){if(n)return t.flags&256?(t.flags&=-257,r=nm(Error(U(422))),Ku(e,t,i,r)):t.memoizedState!==null?(t.child=e.child,t.flags|=128,null):(o=r.fallback,a=t.mode,r=Af({mode:"visible",children:r.children},a,0,null),o=so(o,a,i,null),o.flags|=2,r.return=t,o.return=t,r.sibling=o,t.child=r,t.mode&1&&Ci(t,e.child,null,i),t.child.memoizedState=Ah(i),t.memoizedState=_h,o);if(!(t.mode&1))return Ku(e,t,i,null);if(a.data==="$!"){if(r=a.nextSibling&&a.nextSibling.dataset,r)var s=r.dgst;return r=s,o=Error(U(419)),r=nm(o,r,void 0),Ku(e,t,i,r)}if(s=(i&e.childLanes)!==0,zt||s){if(r=ut,r!==null){switch(i&-i){case 4:a=2;break;case 16:a=8;break;case 64:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:case 4194304:case 8388608:case 16777216:case 33554432:case 67108864:a=32;break;case 536870912:a=268435456;break;default:a=0}a=a&(r.suspendedLanes|i)?0:a,a!==0&&a!==o.retryLane&&(o.retryLane=a,Rr(e,a),Hn(r,e,a,-1))}return dy(),r=nm(Error(U(421))),Ku(e,t,i,r)}return a.data==="$?"?(t.flags|=128,t.child=e.child,t=CC.bind(null,e),a._reactRetry=t,null):(e=o.treeContext,Jt=va(a.nextSibling),tn=t,Be=!0,Mn=null,e!==null&&(pn[fn++]=Sr,pn[fn++]=Or,pn[fn++]=po,Sr=e.id,Or=e.overflow,po=t),t=sy(t,r.children),t.flags|=4096,t)}function X1(e,t,n){e.lanes|=t;var r=e.alternate;r!==null&&(r.lanes|=t),bh(e.return,t,n)}function rm(e,t,n,r,a){var o=e.memoizedState;o===null?e.memoizedState={isBackwards:t,rendering:null,renderingStartTime:0,last:r,tail:n,tailMode:a}:(o.isBackwards=t,o.rendering=null,o.renderingStartTime=0,o.last=r,o.tail=n,o.tailMode=a)}function $9(e,t,n){var r=t.pendingProps,a=r.revealOrder,o=r.tail;if(jt(e,t,r.children,n),r=He.current,r&2)r=r&1|2,t.flags|=128;else{if(e!==null&&e.flags&128)e:for(e=t.child;e!==null;){if(e.tag===13)e.memoizedState!==null&&X1(e,n,t);else if(e.tag===19)X1(e,n,t);else if(e.child!==null){e.child.return=e,e=e.child;continue}if(e===t)break e;for(;e.sibling===null;){if(e.return===null||e.return===t)break e;e=e.return}e.sibling.return=e.return,e=e.sibling}r&=1}if(je(He,r),!(t.mode&1))t.memoizedState=null;else switch(a){case"forwards":for(n=t.child,a=null;n!==null;)e=n.alternate,e!==null&&Jc(e)===null&&(a=n),n=n.sibling;n=a,n===null?(a=t.child,t.child=null):(a=n.sibling,n.sibling=null),rm(t,!1,a,n,o);break;case"backwards":for(n=null,a=t.child,t.child=null;a!==null;){if(e=a.alternate,e!==null&&Jc(e)===null){t.child=a;break}e=a.sibling,a.sibling=n,n=a,a=e}rm(t,!0,n,null,o);break;case"together":rm(t,!1,null,null,void 0);break;default:t.memoizedState=null}return t.child}function _c(e,t){!(t.mode&1)&&e!==null&&(e.alternate=null,t.alternate=null,t.flags|=2)}function Ir(e,t,n){if(e!==null&&(t.dependencies=e.dependencies),mo|=t.lanes,!(n&t.childLanes))return null;if(e!==null&&t.child!==e.child)throw Error(U(153));if(t.child!==null){for(e=t.child,n=wa(e,e.pendingProps),t.child=n,n.return=t;e.sibling!==null;)e=e.sibling,n=n.sibling=wa(e,e.pendingProps),n.return=t;n.sibling=null}return t.child}function mC(e,t,n){switch(t.tag){case 3:j9(t),ki();break;case 5:o9(t);break;case 1:Ut(t.type)&&Vc(t);break;case 4:Jv(t,t.stateNode.containerInfo);break;case 10:var r=t.type._context,a=t.memoizedProps.value;je(Yc,r._currentValue),r._currentValue=a;break;case 13:if(r=t.memoizedState,r!==null)return r.dehydrated!==null?(je(He,He.current&1),t.flags|=128,null):n&t.child.childLanes?N9(e,t,n):(je(He,He.current&1),e=Ir(e,t,n),e!==null?e.sibling:null);je(He,He.current&1);break;case 19:if(r=(n&t.childLanes)!==0,e.flags&128){if(r)return $9(e,t,n);t.flags|=128}if(a=t.memoizedState,a!==null&&(a.rendering=null,a.tail=null,a.lastEffect=null),je(He,He.current),r)break;return null;case 22:case 23:return t.lanes=0,E9(e,t,n)}return Ir(e,t,n)}var M9,Eh,R9,I9;M9=function(e,t){for(var n=t.child;n!==null;){if(n.tag===5||n.tag===6)e.appendChild(n.stateNode);else if(n.tag!==4&&n.child!==null){n.child.return=n,n=n.child;continue}if(n===t)break;for(;n.sibling===null;){if(n.return===null||n.return===t)return;n=n.return}n.sibling.return=n.return,n=n.sibling}};Eh=function(){};R9=function(e,t,n,r){var a=e.memoizedProps;if(a!==r){e=t.stateNode,Ka(lr.current);var o=null;switch(n){case"input":a=Qm(e,a),r=Qm(e,r),o=[];break;case"select":a=Ue({},a,{value:void 0}),r=Ue({},r,{value:void 0}),o=[];break;case"textarea":a=eh(e,a),r=eh(e,r),o=[];break;default:typeof a.onClick!="function"&&typeof r.onClick=="function"&&(e.onclick=Wc)}nh(n,r);var i;n=null;for(u in a)if(!r.hasOwnProperty(u)&&a.hasOwnProperty(u)&&a[u]!=null)if(u==="style"){var s=a[u];for(i in s)s.hasOwnProperty(i)&&(n||(n={}),n[i]="")}else u!=="dangerouslySetInnerHTML"&&u!=="children"&&u!=="suppressContentEditableWarning"&&u!=="suppressHydrationWarning"&&u!=="autoFocus"&&(pl.hasOwnProperty(u)?o||(o=[]):(o=o||[]).push(u,null));for(u in r){var l=r[u];if(s=a!=null?a[u]:void 0,r.hasOwnProperty(u)&&l!==s&&(l!=null||s!=null))if(u==="style")if(s){for(i in s)!s.hasOwnProperty(i)||l&&l.hasOwnProperty(i)||(n||(n={}),n[i]="");for(i in l)l.hasOwnProperty(i)&&s[i]!==l[i]&&(n||(n={}),n[i]=l[i])}else n||(o||(o=[]),o.push(u,n)),n=l;else u==="dangerouslySetInnerHTML"?(l=l?l.__html:void 0,s=s?s.__html:void 0,l!=null&&s!==l&&(o=o||[]).push(u,l)):u==="children"?typeof l!="string"&&typeof l!="number"||(o=o||[]).push(u,""+l):u!=="suppressContentEditableWarning"&&u!=="suppressHydrationWarning"&&(pl.hasOwnProperty(u)?(l!=null&&u==="onScroll"&&Ie("scroll",e),o||s===l||(o=[])):(o=o||[]).push(u,l))}n&&(o=o||[]).push("style",n);var u=o;(t.updateQueue=u)&&(t.flags|=4)}};I9=function(e,t,n,r){n!==r&&(t.flags|=4)};function js(e,t){if(!Be)switch(e.tailMode){case"hidden":t=e.tail;for(var n=null;t!==null;)t.alternate!==null&&(n=t),t=t.sibling;n===null?e.tail=null:n.sibling=null;break;case"collapsed":n=e.tail;for(var r=null;n!==null;)n.alternate!==null&&(r=n),n=n.sibling;r===null?t||e.tail===null?e.tail=null:e.tail.sibling=null:r.sibling=null}}function wt(e){var t=e.alternate!==null&&e.alternate.child===e.child,n=0,r=0;if(t)for(var a=e.child;a!==null;)n|=a.lanes|a.childLanes,r|=a.subtreeFlags&14680064,r|=a.flags&14680064,a.return=e,a=a.sibling;else for(a=e.child;a!==null;)n|=a.lanes|a.childLanes,r|=a.subtreeFlags,r|=a.flags,a.return=e,a=a.sibling;return e.subtreeFlags|=r,e.childLanes=n,t}function hC(e,t,n){var r=t.pendingProps;switch(qv(t),t.tag){case 2:case 16:case 15:case 0:case 11:case 7:case 8:case 12:case 9:case 14:return wt(t),null;case 1:return Ut(t.type)&&qc(),wt(t),null;case 3:return r=t.stateNode,_i(),Fe(Gt),Fe(At),ty(),r.pendingContext&&(r.context=r.pendingContext,r.pendingContext=null),(e===null||e.child===null)&&(qu(t)?t.flags|=4:e===null||e.memoizedState.isDehydrated&&!(t.flags&256)||(t.flags|=1024,Mn!==null&&(Dh(Mn),Mn=null))),Eh(e,t),wt(t),null;case 5:ey(t);var a=Ka(Sl.current);if(n=t.type,e!==null&&t.stateNode!=null)R9(e,t,n,r,a),e.ref!==t.ref&&(t.flags|=512,t.flags|=2097152);else{if(!r){if(t.stateNode===null)throw Error(U(166));return wt(t),null}if(e=Ka(lr.current),qu(t)){r=t.stateNode,n=t.type;var o=t.memoizedProps;switch(r[er]=t,r[bl]=o,e=(t.mode&1)!==0,n){case"dialog":Ie("cancel",r),Ie("close",r);break;case"iframe":case"object":case"embed":Ie("load",r);break;case"video":case"audio":for(a=0;a<Ks.length;a++)Ie(Ks[a],r);break;case"source":Ie("error",r);break;case"img":case"image":case"link":Ie("error",r),Ie("load",r);break;case"details":Ie("toggle",r);break;case"input":a1(r,o),Ie("invalid",r);break;case"select":r._wrapperState={wasMultiple:!!o.multiple},Ie("invalid",r);break;case"textarea":i1(r,o),Ie("invalid",r)}nh(n,o),a=null;for(var i in o)if(o.hasOwnProperty(i)){var s=o[i];i==="children"?typeof s=="string"?r.textContent!==s&&(o.suppressHydrationWarning!==!0&&Wu(r.textContent,s,e),a=["children",s]):typeof s=="number"&&r.textContent!==""+s&&(o.suppressHydrationWarning!==!0&&Wu(r.textContent,s,e),a=["children",""+s]):pl.hasOwnProperty(i)&&s!=null&&i==="onScroll"&&Ie("scroll",r)}switch(n){case"input":Du(r),o1(r,o,!0);break;case"textarea":Du(r),s1(r);break;case"select":case"option":break;default:typeof o.onClick=="function"&&(r.onclick=Wc)}r=a,t.updateQueue=r,r!==null&&(t.flags|=4)}else{i=a.nodeType===9?a:a.ownerDocument,e==="http://www.w3.org/1999/xhtml"&&(e=c5(n)),e==="http://www.w3.org/1999/xhtml"?n==="script"?(e=i.createElement("div"),e.innerHTML="<script><\/script>",e=e.removeChild(e.firstChild)):typeof r.is=="string"?e=i.createElement(n,{is:r.is}):(e=i.createElement(n),n==="select"&&(i=e,r.multiple?i.multiple=!0:r.size&&(i.size=r.size))):e=i.createElementNS(e,n),e[er]=t,e[bl]=r,M9(e,t,!1,!1),t.stateNode=e;e:{switch(i=rh(n,r),n){case"dialog":Ie("cancel",e),Ie("close",e),a=r;break;case"iframe":case"object":case"embed":Ie("load",e),a=r;break;case"video":case"audio":for(a=0;a<Ks.length;a++)Ie(Ks[a],e);a=r;break;case"source":Ie("error",e),a=r;break;case"img":case"image":case"link":Ie("error",e),Ie("load",e),a=r;break;case"details":Ie("toggle",e),a=r;break;case"input":a1(e,r),a=Qm(e,r),Ie("invalid",e);break;case"option":a=r;break;case"select":e._wrapperState={wasMultiple:!!r.multiple},a=Ue({},r,{value:void 0}),Ie("invalid",e);break;case"textarea":i1(e,r),a=eh(e,r),Ie("invalid",e);break;default:a=r}nh(n,a),s=a;for(o in s)if(s.hasOwnProperty(o)){var l=s[o];o==="style"?d5(e,l):o==="dangerouslySetInnerHTML"?(l=l?l.__html:void 0,l!=null&&p5(e,l)):o==="children"?typeof l=="string"?(n!=="textarea"||l!=="")&&fl(e,l):typeof l=="number"&&fl(e,""+l):o!=="suppressContentEditableWarning"&&o!=="suppressHydrationWarning"&&o!=="autoFocus"&&(pl.hasOwnProperty(o)?l!=null&&o==="onScroll"&&Ie("scroll",e):l!=null&&jv(e,o,l,i))}switch(n){case"input":Du(e),o1(e,r,!1);break;case"textarea":Du(e),s1(e);break;case"option":r.value!=null&&e.setAttribute("value",""+Pa(r.value));break;case"select":e.multiple=!!r.multiple,o=r.value,o!=null?ei(e,!!r.multiple,o,!1):r.defaultValue!=null&&ei(e,!!r.multiple,r.defaultValue,!0);break;default:typeof a.onClick=="function"&&(e.onclick=Wc)}switch(n){case"button":case"input":case"select":case"textarea":r=!!r.autoFocus;break e;case"img":r=!0;break e;default:r=!1}}r&&(t.flags|=4)}t.ref!==null&&(t.flags|=512,t.flags|=2097152)}return wt(t),null;case 6:if(e&&t.stateNode!=null)I9(e,t,e.memoizedProps,r);else{if(typeof r!="string"&&t.stateNode===null)throw Error(U(166));if(n=Ka(Sl.current),Ka(lr.current),qu(t)){if(r=t.stateNode,n=t.memoizedProps,r[er]=t,(o=r.nodeValue!==n)&&(e=tn,e!==null))switch(e.tag){case 3:Wu(r.nodeValue,n,(e.mode&1)!==0);break;case 5:e.memoizedProps.suppressHydrationWarning!==!0&&Wu(r.nodeValue,n,(e.mode&1)!==0)}o&&(t.flags|=4)}else r=(n.nodeType===9?n:n.ownerDocument).createTextNode(r),r[er]=t,t.stateNode=r}return wt(t),null;case 13:if(Fe(He),r=t.memoizedState,e===null||e.memoizedState!==null&&e.memoizedState.dehydrated!==null){if(Be&&Jt!==null&&t.mode&1&&!(t.flags&128))e9(),ki(),t.flags|=98560,o=!1;else if(o=qu(t),r!==null&&r.dehydrated!==null){if(e===null){if(!o)throw Error(U(318));if(o=t.memoizedState,o=o!==null?o.dehydrated:null,!o)throw Error(U(317));o[er]=t}else ki(),!(t.flags&128)&&(t.memoizedState=null),t.flags|=4;wt(t),o=!1}else Mn!==null&&(Dh(Mn),Mn=null),o=!0;if(!o)return t.flags&65536?t:null}return t.flags&128?(t.lanes=n,t):(r=r!==null,r!==(e!==null&&e.memoizedState!==null)&&r&&(t.child.flags|=8192,t.mode&1&&(e===null||He.current&1?at===0&&(at=3):dy())),t.updateQueue!==null&&(t.flags|=4),wt(t),null);case 4:return _i(),Eh(e,t),e===null&&xl(t.stateNode.containerInfo),wt(t),null;case 10:return Yv(t.type._context),wt(t),null;case 17:return Ut(t.type)&&qc(),wt(t),null;case 19:if(Fe(He),o=t.memoizedState,o===null)return wt(t),null;if(r=(t.flags&128)!==0,i=o.rendering,i===null)if(r)js(o,!1);else{if(at!==0||e!==null&&e.flags&128)for(e=t.child;e!==null;){if(i=Jc(e),i!==null){for(t.flags|=128,js(o,!1),r=i.updateQueue,r!==null&&(t.updateQueue=r,t.flags|=4),t.subtreeFlags=0,r=n,n=t.child;n!==null;)o=n,e=r,o.flags&=14680066,i=o.alternate,i===null?(o.childLanes=0,o.lanes=e,o.child=null,o.subtreeFlags=0,o.memoizedProps=null,o.memoizedState=null,o.updateQueue=null,o.dependencies=null,o.stateNode=null):(o.childLanes=i.childLanes,o.lanes=i.lanes,o.child=i.child,o.subtreeFlags=0,o.deletions=null,o.memoizedProps=i.memoizedProps,o.memoizedState=i.memoizedState,o.updateQueue=i.updateQueue,o.type=i.type,e=i.dependencies,o.dependencies=e===null?null:{lanes:e.lanes,firstContext:e.firstContext}),n=n.sibling;return je(He,He.current&1|2),t.child}e=e.sibling}o.tail!==null&&Ke()>Ei&&(t.flags|=128,r=!0,js(o,!1),t.lanes=4194304)}else{if(!r)if(e=Jc(i),e!==null){if(t.flags|=128,r=!0,n=e.updateQueue,n!==null&&(t.updateQueue=n,t.flags|=4),js(o,!0),o.tail===null&&o.tailMode==="hidden"&&!i.alternate&&!Be)return wt(t),null}else 2*Ke()-o.renderingStartTime>Ei&&n!==1073741824&&(t.flags|=128,r=!0,js(o,!1),t.lanes=4194304);o.isBackwards?(i.sibling=t.child,t.child=i):(n=o.last,n!==null?n.sibling=i:t.child=i,o.last=i)}return o.tail!==null?(t=o.tail,o.rendering=t,o.tail=t.sibling,o.renderingStartTime=Ke(),t.sibling=null,n=He.current,je(He,r?n&1|2:n&1),t):(wt(t),null);case 22:case 23:return fy(),r=t.memoizedState!==null,e!==null&&e.memoizedState!==null!==r&&(t.flags|=8192),r&&t.mode&1?Xt&1073741824&&(wt(t),t.subtreeFlags&6&&(t.flags|=8192)):wt(t),null;case 24:return null;case 25:return null}throw Error(U(156,t.tag))}function vC(e,t){switch(qv(t),t.tag){case 1:return Ut(t.type)&&qc(),e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 3:return _i(),Fe(Gt),Fe(At),ty(),e=t.flags,e&65536&&!(e&128)?(t.flags=e&-65537|128,t):null;case 5:return ey(t),null;case 13:if(Fe(He),e=t.memoizedState,e!==null&&e.dehydrated!==null){if(t.alternate===null)throw Error(U(340));ki()}return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 19:return Fe(He),null;case 4:return _i(),null;case 10:return Yv(t.type._context),null;case 22:case 23:return fy(),null;case 24:return null;default:return null}}var Xu=!1,Ot=!1,yC=typeof WeakSet=="function"?WeakSet:Set,Y=null;function Xo(e,t){var n=e.ref;if(n!==null)if(typeof n=="function")try{n(null)}catch(r){qe(e,t,r)}else n.current=null}function Th(e,t,n){try{n()}catch(r){qe(e,t,r)}}var Y1=!1;function gC(e,t){if(dh=Hc,e=z5(),Uv(e)){if("selectionStart"in e)var n={start:e.selectionStart,end:e.selectionEnd};else e:{n=(n=e.ownerDocument)&&n.defaultView||window;var r=n.getSelection&&n.getSelection();if(r&&r.rangeCount!==0){n=r.anchorNode;var a=r.anchorOffset,o=r.focusNode;r=r.focusOffset;try{n.nodeType,o.nodeType}catch{n=null;break e}var i=0,s=-1,l=-1,u=0,p=0,c=e,f=null;t:for(;;){for(var m;c!==n||a!==0&&c.nodeType!==3||(s=i+a),c!==o||r!==0&&c.nodeType!==3||(l=i+r),c.nodeType===3&&(i+=c.nodeValue.length),(m=c.firstChild)!==null;)f=c,c=m;for(;;){if(c===e)break t;if(f===n&&++u===a&&(s=i),f===o&&++p===r&&(l=i),(m=c.nextSibling)!==null)break;c=f,f=c.parentNode}c=m}n=s===-1||l===-1?null:{start:s,end:l}}else n=null}n=n||{start:0,end:0}}else n=null;for(mh={focusedElem:e,selectionRange:n},Hc=!1,Y=t;Y!==null;)if(t=Y,e=t.child,(t.subtreeFlags&1028)!==0&&e!==null)e.return=t,Y=e;else for(;Y!==null;){t=Y;try{var h=t.alternate;if(t.flags&1024)switch(t.tag){case 0:case 11:case 15:break;case 1:if(h!==null){var d=h.memoizedProps,g=h.memoizedState,v=t.stateNode,y=v.getSnapshotBeforeUpdate(t.elementType===t.type?d:_n(t.type,d),g);v.__reactInternalSnapshotBeforeUpdate=y}break;case 3:var x=t.stateNode.containerInfo;x.nodeType===1?x.textContent="":x.nodeType===9&&x.documentElement&&x.removeChild(x.documentElement);break;case 5:case 6:case 4:case 17:break;default:throw Error(U(163))}}catch(P){qe(t,t.return,P)}if(e=t.sibling,e!==null){e.return=t.return,Y=e;break}Y=t.return}return h=Y1,Y1=!1,h}function rl(e,t,n){var r=t.updateQueue;if(r=r!==null?r.lastEffect:null,r!==null){var a=r=r.next;do{if((a.tag&e)===e){var o=a.destroy;a.destroy=void 0,o!==void 0&&Th(t,n,o)}a=a.next}while(a!==r)}}function Cf(e,t){if(t=t.updateQueue,t=t!==null?t.lastEffect:null,t!==null){var n=t=t.next;do{if((n.tag&e)===e){var r=n.create;n.destroy=r()}n=n.next}while(n!==t)}}function jh(e){var t=e.ref;if(t!==null){var n=e.stateNode;switch(e.tag){case 5:e=n;break;default:e=n}typeof t=="function"?t(e):t.current=e}}function D9(e){var t=e.alternate;t!==null&&(e.alternate=null,D9(t)),e.child=null,e.deletions=null,e.sibling=null,e.tag===5&&(t=e.stateNode,t!==null&&(delete t[er],delete t[bl],delete t[yh],delete t[eC],delete t[tC])),e.stateNode=null,e.return=null,e.dependencies=null,e.memoizedProps=null,e.memoizedState=null,e.pendingProps=null,e.stateNode=null,e.updateQueue=null}function L9(e){return e.tag===5||e.tag===3||e.tag===4}function Q1(e){e:for(;;){for(;e.sibling===null;){if(e.return===null||L9(e.return))return null;e=e.return}for(e.sibling.return=e.return,e=e.sibling;e.tag!==5&&e.tag!==6&&e.tag!==18;){if(e.flags&2||e.child===null||e.tag===4)continue e;e.child.return=e,e=e.child}if(!(e.flags&2))return e.stateNode}}function Nh(e,t,n){var r=e.tag;if(r===5||r===6)e=e.stateNode,t?n.nodeType===8?n.parentNode.insertBefore(e,t):n.insertBefore(e,t):(n.nodeType===8?(t=n.parentNode,t.insertBefore(e,n)):(t=n,t.appendChild(e)),n=n._reactRootContainer,n!=null||t.onclick!==null||(t.onclick=Wc));else if(r!==4&&(e=e.child,e!==null))for(Nh(e,t,n),e=e.sibling;e!==null;)Nh(e,t,n),e=e.sibling}function $h(e,t,n){var r=e.tag;if(r===5||r===6)e=e.stateNode,t?n.insertBefore(e,t):n.appendChild(e);else if(r!==4&&(e=e.child,e!==null))for($h(e,t,n),e=e.sibling;e!==null;)$h(e,t,n),e=e.sibling}var ft=null,jn=!1;function qr(e,t,n){for(n=n.child;n!==null;)F9(e,t,n),n=n.sibling}function F9(e,t,n){if(sr&&typeof sr.onCommitFiberUnmount=="function")try{sr.onCommitFiberUnmount(gf,n)}catch{}switch(n.tag){case 5:Ot||Xo(n,t);case 6:var r=ft,a=jn;ft=null,qr(e,t,n),ft=r,jn=a,ft!==null&&(jn?(e=ft,n=n.stateNode,e.nodeType===8?e.parentNode.removeChild(n):e.removeChild(n)):ft.removeChild(n.stateNode));break;case 18:ft!==null&&(jn?(e=ft,n=n.stateNode,e.nodeType===8?Yd(e.parentNode,n):e.nodeType===1&&Yd(e,n),vl(e)):Yd(ft,n.stateNode));break;case 4:r=ft,a=jn,ft=n.stateNode.containerInfo,jn=!0,qr(e,t,n),ft=r,jn=a;break;case 0:case 11:case 14:case 15:if(!Ot&&(r=n.updateQueue,r!==null&&(r=r.lastEffect,r!==null))){a=r=r.next;do{var o=a,i=o.destroy;o=o.tag,i!==void 0&&(o&2||o&4)&&Th(n,t,i),a=a.next}while(a!==r)}qr(e,t,n);break;case 1:if(!Ot&&(Xo(n,t),r=n.stateNode,typeof r.componentWillUnmount=="function"))try{r.props=n.memoizedProps,r.state=n.memoizedState,r.componentWillUnmount()}catch(s){qe(n,t,s)}qr(e,t,n);break;case 21:qr(e,t,n);break;case 22:n.mode&1?(Ot=(r=Ot)||n.memoizedState!==null,qr(e,t,n),Ot=r):qr(e,t,n);break;default:qr(e,t,n)}}function Z1(e){var t=e.updateQueue;if(t!==null){e.updateQueue=null;var n=e.stateNode;n===null&&(n=e.stateNode=new yC),t.forEach(function(r){var a=_C.bind(null,e,r);n.has(r)||(n.add(r),r.then(a,a))})}}function On(e,t){var n=t.deletions;if(n!==null)for(var r=0;r<n.length;r++){var a=n[r];try{var o=e,i=t,s=i;e:for(;s!==null;){switch(s.tag){case 5:ft=s.stateNode,jn=!1;break e;case 3:ft=s.stateNode.containerInfo,jn=!0;break e;case 4:ft=s.stateNode.containerInfo,jn=!0;break e}s=s.return}if(ft===null)throw Error(U(160));F9(o,i,a),ft=null,jn=!1;var l=a.alternate;l!==null&&(l.return=null),a.return=null}catch(u){qe(a,t,u)}}if(t.subtreeFlags&12854)for(t=t.child;t!==null;)B9(t,e),t=t.sibling}function B9(e,t){var n=e.alternate,r=e.flags;switch(e.tag){case 0:case 11:case 14:case 15:if(On(t,e),Kn(e),r&4){try{rl(3,e,e.return),Cf(3,e)}catch(d){qe(e,e.return,d)}try{rl(5,e,e.return)}catch(d){qe(e,e.return,d)}}break;case 1:On(t,e),Kn(e),r&512&&n!==null&&Xo(n,n.return);break;case 5:if(On(t,e),Kn(e),r&512&&n!==null&&Xo(n,n.return),e.flags&32){var a=e.stateNode;try{fl(a,"")}catch(d){qe(e,e.return,d)}}if(r&4&&(a=e.stateNode,a!=null)){var o=e.memoizedProps,i=n!==null?n.memoizedProps:o,s=e.type,l=e.updateQueue;if(e.updateQueue=null,l!==null)try{s==="input"&&o.type==="radio"&&o.name!=null&&l5(a,o),rh(s,i);var u=rh(s,o);for(i=0;i<l.length;i+=2){var p=l[i],c=l[i+1];p==="style"?d5(a,c):p==="dangerouslySetInnerHTML"?p5(a,c):p==="children"?fl(a,c):jv(a,p,c,u)}switch(s){case"input":Zm(a,o);break;case"textarea":u5(a,o);break;case"select":var f=a._wrapperState.wasMultiple;a._wrapperState.wasMultiple=!!o.multiple;var m=o.value;m!=null?ei(a,!!o.multiple,m,!1):f!==!!o.multiple&&(o.defaultValue!=null?ei(a,!!o.multiple,o.defaultValue,!0):ei(a,!!o.multiple,o.multiple?[]:"",!1))}a[bl]=o}catch(d){qe(e,e.return,d)}}break;case 6:if(On(t,e),Kn(e),r&4){if(e.stateNode===null)throw Error(U(162));a=e.stateNode,o=e.memoizedProps;try{a.nodeValue=o}catch(d){qe(e,e.return,d)}}break;case 3:if(On(t,e),Kn(e),r&4&&n!==null&&n.memoizedState.isDehydrated)try{vl(t.containerInfo)}catch(d){qe(e,e.return,d)}break;case 4:On(t,e),Kn(e);break;case 13:On(t,e),Kn(e),a=e.child,a.flags&8192&&(o=a.memoizedState!==null,a.stateNode.isHidden=o,!o||a.alternate!==null&&a.alternate.memoizedState!==null||(cy=Ke())),r&4&&Z1(e);break;case 22:if(p=n!==null&&n.memoizedState!==null,e.mode&1?(Ot=(u=Ot)||p,On(t,e),Ot=u):On(t,e),Kn(e),r&8192){if(u=e.memoizedState!==null,(e.stateNode.isHidden=u)&&!p&&e.mode&1)for(Y=e,p=e.child;p!==null;){for(c=Y=p;Y!==null;){switch(f=Y,m=f.child,f.tag){case 0:case 11:case 14:case 15:rl(4,f,f.return);break;case 1:Xo(f,f.return);var h=f.stateNode;if(typeof h.componentWillUnmount=="function"){r=f,n=f.return;try{t=r,h.props=t.memoizedProps,h.state=t.memoizedState,h.componentWillUnmount()}catch(d){qe(r,n,d)}}break;case 5:Xo(f,f.return);break;case 22:if(f.memoizedState!==null){ex(c);continue}}m!==null?(m.return=f,Y=m):ex(c)}p=p.sibling}e:for(p=null,c=e;;){if(c.tag===5){if(p===null){p=c;try{a=c.stateNode,u?(o=a.style,typeof o.setProperty=="function"?o.setProperty("display","none","important"):o.display="none"):(s=c.stateNode,l=c.memoizedProps.style,i=l!=null&&l.hasOwnProperty("display")?l.display:null,s.style.display=f5("display",i))}catch(d){qe(e,e.return,d)}}}else if(c.tag===6){if(p===null)try{c.stateNode.nodeValue=u?"":c.memoizedProps}catch(d){qe(e,e.return,d)}}else if((c.tag!==22&&c.tag!==23||c.memoizedState===null||c===e)&&c.child!==null){c.child.return=c,c=c.child;continue}if(c===e)break e;for(;c.sibling===null;){if(c.return===null||c.return===e)break e;p===c&&(p=null),c=c.return}p===c&&(p=null),c.sibling.return=c.return,c=c.sibling}}break;case 19:On(t,e),Kn(e),r&4&&Z1(e);break;case 21:break;default:On(t,e),Kn(e)}}function Kn(e){var t=e.flags;if(t&2){try{e:{for(var n=e.return;n!==null;){if(L9(n)){var r=n;break e}n=n.return}throw Error(U(160))}switch(r.tag){case 5:var a=r.stateNode;r.flags&32&&(fl(a,""),r.flags&=-33);var o=Q1(e);$h(e,o,a);break;case 3:case 4:var i=r.stateNode.containerInfo,s=Q1(e);Nh(e,s,i);break;default:throw Error(U(161))}}catch(l){qe(e,e.return,l)}e.flags&=-3}t&4096&&(e.flags&=-4097)}function xC(e,t,n){Y=e,z9(e)}function z9(e,t,n){for(var r=(e.mode&1)!==0;Y!==null;){var a=Y,o=a.child;if(a.tag===22&&r){var i=a.memoizedState!==null||Xu;if(!i){var s=a.alternate,l=s!==null&&s.memoizedState!==null||Ot;s=Xu;var u=Ot;if(Xu=i,(Ot=l)&&!u)for(Y=a;Y!==null;)i=Y,l=i.child,i.tag===22&&i.memoizedState!==null?tx(a):l!==null?(l.return=i,Y=l):tx(a);for(;o!==null;)Y=o,z9(o),o=o.sibling;Y=a,Xu=s,Ot=u}J1(e)}else a.subtreeFlags&8772&&o!==null?(o.return=a,Y=o):J1(e)}}function J1(e){for(;Y!==null;){var t=Y;if(t.flags&8772){var n=t.alternate;try{if(t.flags&8772)switch(t.tag){case 0:case 11:case 15:Ot||Cf(5,t);break;case 1:var r=t.stateNode;if(t.flags&4&&!Ot)if(n===null)r.componentDidMount();else{var a=t.elementType===t.type?n.memoizedProps:_n(t.type,n.memoizedProps);r.componentDidUpdate(a,n.memoizedState,r.__reactInternalSnapshotBeforeUpdate)}var o=t.updateQueue;o!==null&&D1(t,o,r);break;case 3:var i=t.updateQueue;if(i!==null){if(n=null,t.child!==null)switch(t.child.tag){case 5:n=t.child.stateNode;break;case 1:n=t.child.stateNode}D1(t,i,n)}break;case 5:var s=t.stateNode;if(n===null&&t.flags&4){n=s;var l=t.memoizedProps;switch(t.type){case"button":case"input":case"select":case"textarea":l.autoFocus&&n.focus();break;case"img":l.src&&(n.src=l.src)}}break;case 6:break;case 4:break;case 12:break;case 13:if(t.memoizedState===null){var u=t.alternate;if(u!==null){var p=u.memoizedState;if(p!==null){var c=p.dehydrated;c!==null&&vl(c)}}}break;case 19:case 17:case 21:case 22:case 23:case 25:break;default:throw Error(U(163))}Ot||t.flags&512&&jh(t)}catch(f){qe(t,t.return,f)}}if(t===e){Y=null;break}if(n=t.sibling,n!==null){n.return=t.return,Y=n;break}Y=t.return}}function ex(e){for(;Y!==null;){var t=Y;if(t===e){Y=null;break}var n=t.sibling;if(n!==null){n.return=t.return,Y=n;break}Y=t.return}}function tx(e){for(;Y!==null;){var t=Y;try{switch(t.tag){case 0:case 11:case 15:var n=t.return;try{Cf(4,t)}catch(l){qe(t,n,l)}break;case 1:var r=t.stateNode;if(typeof r.componentDidMount=="function"){var a=t.return;try{r.componentDidMount()}catch(l){qe(t,a,l)}}var o=t.return;try{jh(t)}catch(l){qe(t,o,l)}break;case 5:var i=t.return;try{jh(t)}catch(l){qe(t,i,l)}}}catch(l){qe(t,t.return,l)}if(t===e){Y=null;break}var s=t.sibling;if(s!==null){s.return=t.return,Y=s;break}Y=t.return}}var wC=Math.ceil,np=Hr.ReactCurrentDispatcher,ly=Hr.ReactCurrentOwner,hn=Hr.ReactCurrentBatchConfig,de=0,ut=null,Je=null,vt=0,Xt=0,Yo=Aa(0),at=0,_l=null,mo=0,_f=0,uy=0,al=null,Ft=null,cy=0,Ei=1/0,wr=null,rp=!1,Mh=null,ga=null,Yu=!1,ca=null,ap=0,ol=0,Rh=null,Ac=-1,Ec=0;function Rt(){return de&6?Ke():Ac!==-1?Ac:Ac=Ke()}function xa(e){return e.mode&1?de&2&&vt!==0?vt&-vt:rC.transition!==null?(Ec===0&&(Ec=k5()),Ec):(e=Oe,e!==0||(e=window.event,e=e===void 0?16:N5(e.type)),e):1}function Hn(e,t,n,r){if(50<ol)throw ol=0,Rh=null,Error(U(185));gu(e,n,r),(!(de&2)||e!==ut)&&(e===ut&&(!(de&2)&&(_f|=n),at===4&&na(e,vt)),Wt(e,r),n===1&&de===0&&!(t.mode&1)&&(Ei=Ke()+500,Sf&&Ea()))}function Wt(e,t){var n=e.callbackNode;rk(e,t);var r=zc(e,e===ut?vt:0);if(r===0)n!==null&&c1(n),e.callbackNode=null,e.callbackPriority=0;else if(t=r&-r,e.callbackPriority!==t){if(n!=null&&c1(n),t===1)e.tag===0?nC(nx.bind(null,e)):Q5(nx.bind(null,e)),Zk(function(){!(de&6)&&Ea()}),n=null;else{switch(C5(r)){case 1:n=Iv;break;case 4:n=S5;break;case 16:n=Bc;break;case 536870912:n=O5;break;default:n=Bc}n=X9(n,H9.bind(null,e))}e.callbackPriority=t,e.callbackNode=n}}function H9(e,t){if(Ac=-1,Ec=0,de&6)throw Error(U(327));var n=e.callbackNode;if(oi()&&e.callbackNode!==n)return null;var r=zc(e,e===ut?vt:0);if(r===0)return null;if(r&30||r&e.expiredLanes||t)t=op(e,r);else{t=r;var a=de;de|=2;var o=U9();(ut!==e||vt!==t)&&(wr=null,Ei=Ke()+500,io(e,t));do try{SC();break}catch(s){G9(e,s)}while(!0);Xv(),np.current=o,de=a,Je!==null?t=0:(ut=null,vt=0,t=at)}if(t!==0){if(t===2&&(a=lh(e),a!==0&&(r=a,t=Ih(e,a))),t===1)throw n=_l,io(e,0),na(e,r),Wt(e,Ke()),n;if(t===6)na(e,r);else{if(a=e.current.alternate,!(r&30)&&!bC(a)&&(t=op(e,r),t===2&&(o=lh(e),o!==0&&(r=o,t=Ih(e,o))),t===1))throw n=_l,io(e,0),na(e,r),Wt(e,Ke()),n;switch(e.finishedWork=a,e.finishedLanes=r,t){case 0:case 1:throw Error(U(345));case 2:Ha(e,Ft,wr);break;case 3:if(na(e,r),(r&130023424)===r&&(t=cy+500-Ke(),10<t)){if(zc(e,0)!==0)break;if(a=e.suspendedLanes,(a&r)!==r){Rt(),e.pingedLanes|=e.suspendedLanes&a;break}e.timeoutHandle=vh(Ha.bind(null,e,Ft,wr),t);break}Ha(e,Ft,wr);break;case 4:if(na(e,r),(r&4194240)===r)break;for(t=e.eventTimes,a=-1;0<r;){var i=31-zn(r);o=1<<i,i=t[i],i>a&&(a=i),r&=~o}if(r=a,r=Ke()-r,r=(120>r?120:480>r?480:1080>r?1080:1920>r?1920:3e3>r?3e3:4320>r?4320:1960*wC(r/1960))-r,10<r){e.timeoutHandle=vh(Ha.bind(null,e,Ft,wr),r);break}Ha(e,Ft,wr);break;case 5:Ha(e,Ft,wr);break;default:throw Error(U(329))}}}return Wt(e,Ke()),e.callbackNode===n?H9.bind(null,e):null}function Ih(e,t){var n=al;return e.current.memoizedState.isDehydrated&&(io(e,t).flags|=256),e=op(e,t),e!==2&&(t=Ft,Ft=n,t!==null&&Dh(t)),e}function Dh(e){Ft===null?Ft=e:Ft.push.apply(Ft,e)}function bC(e){for(var t=e;;){if(t.flags&16384){var n=t.updateQueue;if(n!==null&&(n=n.stores,n!==null))for(var r=0;r<n.length;r++){var a=n[r],o=a.getSnapshot;a=a.value;try{if(!Gn(o(),a))return!1}catch{return!1}}}if(n=t.child,t.subtreeFlags&16384&&n!==null)n.return=t,t=n;else{if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return!0;t=t.return}t.sibling.return=t.return,t=t.sibling}}return!0}function na(e,t){for(t&=~uy,t&=~_f,e.suspendedLanes|=t,e.pingedLanes&=~t,e=e.expirationTimes;0<t;){var n=31-zn(t),r=1<<n;e[n]=-1,t&=~r}}function nx(e){if(de&6)throw Error(U(327));oi();var t=zc(e,0);if(!(t&1))return Wt(e,Ke()),null;var n=op(e,t);if(e.tag!==0&&n===2){var r=lh(e);r!==0&&(t=r,n=Ih(e,r))}if(n===1)throw n=_l,io(e,0),na(e,t),Wt(e,Ke()),n;if(n===6)throw Error(U(345));return e.finishedWork=e.current.alternate,e.finishedLanes=t,Ha(e,Ft,wr),Wt(e,Ke()),null}function py(e,t){var n=de;de|=1;try{return e(t)}finally{de=n,de===0&&(Ei=Ke()+500,Sf&&Ea())}}function ho(e){ca!==null&&ca.tag===0&&!(de&6)&&oi();var t=de;de|=1;var n=hn.transition,r=Oe;try{if(hn.transition=null,Oe=1,e)return e()}finally{Oe=r,hn.transition=n,de=t,!(de&6)&&Ea()}}function fy(){Xt=Yo.current,Fe(Yo)}function io(e,t){e.finishedWork=null,e.finishedLanes=0;var n=e.timeoutHandle;if(n!==-1&&(e.timeoutHandle=-1,Qk(n)),Je!==null)for(n=Je.return;n!==null;){var r=n;switch(qv(r),r.tag){case 1:r=r.type.childContextTypes,r!=null&&qc();break;case 3:_i(),Fe(Gt),Fe(At),ty();break;case 5:ey(r);break;case 4:_i();break;case 13:Fe(He);break;case 19:Fe(He);break;case 10:Yv(r.type._context);break;case 22:case 23:fy()}n=n.return}if(ut=e,Je=e=wa(e.current,null),vt=Xt=t,at=0,_l=null,uy=_f=mo=0,Ft=al=null,Va!==null){for(t=0;t<Va.length;t++)if(n=Va[t],r=n.interleaved,r!==null){n.interleaved=null;var a=r.next,o=n.pending;if(o!==null){var i=o.next;o.next=a,r.next=i}n.pending=r}Va=null}return e}function G9(e,t){do{var n=Je;try{if(Xv(),kc.current=tp,ep){for(var r=Ge.memoizedState;r!==null;){var a=r.queue;a!==null&&(a.pending=null),r=r.next}ep=!1}if(fo=0,lt=rt=Ge=null,nl=!1,Ol=0,ly.current=null,n===null||n.return===null){at=1,_l=t,Je=null;break}e:{var o=e,i=n.return,s=n,l=t;if(t=vt,s.flags|=32768,l!==null&&typeof l=="object"&&typeof l.then=="function"){var u=l,p=s,c=p.tag;if(!(p.mode&1)&&(c===0||c===11||c===15)){var f=p.alternate;f?(p.updateQueue=f.updateQueue,p.memoizedState=f.memoizedState,p.lanes=f.lanes):(p.updateQueue=null,p.memoizedState=null)}var m=G1(i);if(m!==null){m.flags&=-257,U1(m,i,s,o,t),m.mode&1&&H1(o,u,t),t=m,l=u;var h=t.updateQueue;if(h===null){var d=new Set;d.add(l),t.updateQueue=d}else h.add(l);break e}else{if(!(t&1)){H1(o,u,t),dy();break e}l=Error(U(426))}}else if(Be&&s.mode&1){var g=G1(i);if(g!==null){!(g.flags&65536)&&(g.flags|=256),U1(g,i,s,o,t),Vv(Ai(l,s));break e}}o=l=Ai(l,s),at!==4&&(at=2),al===null?al=[o]:al.push(o),o=i;do{switch(o.tag){case 3:o.flags|=65536,t&=-t,o.lanes|=t;var v=C9(o,l,t);I1(o,v);break e;case 1:s=l;var y=o.type,x=o.stateNode;if(!(o.flags&128)&&(typeof y.getDerivedStateFromError=="function"||x!==null&&typeof x.componentDidCatch=="function"&&(ga===null||!ga.has(x)))){o.flags|=65536,t&=-t,o.lanes|=t;var P=_9(o,s,t);I1(o,P);break e}}o=o.return}while(o!==null)}q9(n)}catch(w){t=w,Je===n&&n!==null&&(Je=n=n.return);continue}break}while(!0)}function U9(){var e=np.current;return np.current=tp,e===null?tp:e}function dy(){(at===0||at===3||at===2)&&(at=4),ut===null||!(mo&268435455)&&!(_f&268435455)||na(ut,vt)}function op(e,t){var n=de;de|=2;var r=U9();(ut!==e||vt!==t)&&(wr=null,io(e,t));do try{PC();break}catch(a){G9(e,a)}while(!0);if(Xv(),de=n,np.current=r,Je!==null)throw Error(U(261));return ut=null,vt=0,at}function PC(){for(;Je!==null;)W9(Je)}function SC(){for(;Je!==null&&!KO();)W9(Je)}function W9(e){var t=K9(e.alternate,e,Xt);e.memoizedProps=e.pendingProps,t===null?q9(e):Je=t,ly.current=null}function q9(e){var t=e;do{var n=t.alternate;if(e=t.return,t.flags&32768){if(n=vC(n,t),n!==null){n.flags&=32767,Je=n;return}if(e!==null)e.flags|=32768,e.subtreeFlags=0,e.deletions=null;else{at=6,Je=null;return}}else if(n=hC(n,t,Xt),n!==null){Je=n;return}if(t=t.sibling,t!==null){Je=t;return}Je=t=e}while(t!==null);at===0&&(at=5)}function Ha(e,t,n){var r=Oe,a=hn.transition;try{hn.transition=null,Oe=1,OC(e,t,n,r)}finally{hn.transition=a,Oe=r}return null}function OC(e,t,n,r){do oi();while(ca!==null);if(de&6)throw Error(U(327));n=e.finishedWork;var a=e.finishedLanes;if(n===null)return null;if(e.finishedWork=null,e.finishedLanes=0,n===e.current)throw Error(U(177));e.callbackNode=null,e.callbackPriority=0;var o=n.lanes|n.childLanes;if(ak(e,o),e===ut&&(Je=ut=null,vt=0),!(n.subtreeFlags&2064)&&!(n.flags&2064)||Yu||(Yu=!0,X9(Bc,function(){return oi(),null})),o=(n.flags&15990)!==0,n.subtreeFlags&15990||o){o=hn.transition,hn.transition=null;var i=Oe;Oe=1;var s=de;de|=4,ly.current=null,gC(e,n),B9(n,e),Uk(mh),Hc=!!dh,mh=dh=null,e.current=n,xC(n),XO(),de=s,Oe=i,hn.transition=o}else e.current=n;if(Yu&&(Yu=!1,ca=e,ap=a),o=e.pendingLanes,o===0&&(ga=null),ZO(n.stateNode),Wt(e,Ke()),t!==null)for(r=e.onRecoverableError,n=0;n<t.length;n++)a=t[n],r(a.value,{componentStack:a.stack,digest:a.digest});if(rp)throw rp=!1,e=Mh,Mh=null,e;return ap&1&&e.tag!==0&&oi(),o=e.pendingLanes,o&1?e===Rh?ol++:(ol=0,Rh=e):ol=0,Ea(),null}function oi(){if(ca!==null){var e=C5(ap),t=hn.transition,n=Oe;try{if(hn.transition=null,Oe=16>e?16:e,ca===null)var r=!1;else{if(e=ca,ca=null,ap=0,de&6)throw Error(U(331));var a=de;for(de|=4,Y=e.current;Y!==null;){var o=Y,i=o.child;if(Y.flags&16){var s=o.deletions;if(s!==null){for(var l=0;l<s.length;l++){var u=s[l];for(Y=u;Y!==null;){var p=Y;switch(p.tag){case 0:case 11:case 15:rl(8,p,o)}var c=p.child;if(c!==null)c.return=p,Y=c;else for(;Y!==null;){p=Y;var f=p.sibling,m=p.return;if(D9(p),p===u){Y=null;break}if(f!==null){f.return=m,Y=f;break}Y=m}}}var h=o.alternate;if(h!==null){var d=h.child;if(d!==null){h.child=null;do{var g=d.sibling;d.sibling=null,d=g}while(d!==null)}}Y=o}}if(o.subtreeFlags&2064&&i!==null)i.return=o,Y=i;else e:for(;Y!==null;){if(o=Y,o.flags&2048)switch(o.tag){case 0:case 11:case 15:rl(9,o,o.return)}var v=o.sibling;if(v!==null){v.return=o.return,Y=v;break e}Y=o.return}}var y=e.current;for(Y=y;Y!==null;){i=Y;var x=i.child;if(i.subtreeFlags&2064&&x!==null)x.return=i,Y=x;else e:for(i=y;Y!==null;){if(s=Y,s.flags&2048)try{switch(s.tag){case 0:case 11:case 15:Cf(9,s)}}catch(w){qe(s,s.return,w)}if(s===i){Y=null;break e}var P=s.sibling;if(P!==null){P.return=s.return,Y=P;break e}Y=s.return}}if(de=a,Ea(),sr&&typeof sr.onPostCommitFiberRoot=="function")try{sr.onPostCommitFiberRoot(gf,e)}catch{}r=!0}return r}finally{Oe=n,hn.transition=t}}return!1}function rx(e,t,n){t=Ai(n,t),t=C9(e,t,1),e=ya(e,t,1),t=Rt(),e!==null&&(gu(e,1,t),Wt(e,t))}function qe(e,t,n){if(e.tag===3)rx(e,e,n);else for(;t!==null;){if(t.tag===3){rx(t,e,n);break}else if(t.tag===1){var r=t.stateNode;if(typeof t.type.getDerivedStateFromError=="function"||typeof r.componentDidCatch=="function"&&(ga===null||!ga.has(r))){e=Ai(n,e),e=_9(t,e,1),t=ya(t,e,1),e=Rt(),t!==null&&(gu(t,1,e),Wt(t,e));break}}t=t.return}}function kC(e,t,n){var r=e.pingCache;r!==null&&r.delete(t),t=Rt(),e.pingedLanes|=e.suspendedLanes&n,ut===e&&(vt&n)===n&&(at===4||at===3&&(vt&130023424)===vt&&500>Ke()-cy?io(e,0):uy|=n),Wt(e,t)}function V9(e,t){t===0&&(e.mode&1?(t=Bu,Bu<<=1,!(Bu&130023424)&&(Bu=4194304)):t=1);var n=Rt();e=Rr(e,t),e!==null&&(gu(e,t,n),Wt(e,n))}function CC(e){var t=e.memoizedState,n=0;t!==null&&(n=t.retryLane),V9(e,n)}function _C(e,t){var n=0;switch(e.tag){case 13:var r=e.stateNode,a=e.memoizedState;a!==null&&(n=a.retryLane);break;case 19:r=e.stateNode;break;default:throw Error(U(314))}r!==null&&r.delete(t),V9(e,n)}var K9;K9=function(e,t,n){if(e!==null)if(e.memoizedProps!==t.pendingProps||Gt.current)zt=!0;else{if(!(e.lanes&n)&&!(t.flags&128))return zt=!1,mC(e,t,n);zt=!!(e.flags&131072)}else zt=!1,Be&&t.flags&1048576&&Z5(t,Xc,t.index);switch(t.lanes=0,t.tag){case 2:var r=t.type;_c(e,t),e=t.pendingProps;var a=Oi(t,At.current);ai(t,n),a=ry(null,t,r,e,a,n);var o=ay();return t.flags|=1,typeof a=="object"&&a!==null&&typeof a.render=="function"&&a.$$typeof===void 0?(t.tag=1,t.memoizedState=null,t.updateQueue=null,Ut(r)?(o=!0,Vc(t)):o=!1,t.memoizedState=a.state!==null&&a.state!==void 0?a.state:null,Zv(t),a.updater=kf,t.stateNode=a,a._reactInternals=t,Sh(t,r,e,n),t=Ch(null,t,r,!0,o,n)):(t.tag=0,Be&&o&&Wv(t),jt(null,t,a,n),t=t.child),t;case 16:r=t.elementType;e:{switch(_c(e,t),e=t.pendingProps,a=r._init,r=a(r._payload),t.type=r,a=t.tag=EC(r),e=_n(r,e),a){case 0:t=kh(null,t,r,e,n);break e;case 1:t=V1(null,t,r,e,n);break e;case 11:t=W1(null,t,r,e,n);break e;case 14:t=q1(null,t,r,_n(r.type,e),n);break e}throw Error(U(306,r,""))}return t;case 0:return r=t.type,a=t.pendingProps,a=t.elementType===r?a:_n(r,a),kh(e,t,r,a,n);case 1:return r=t.type,a=t.pendingProps,a=t.elementType===r?a:_n(r,a),V1(e,t,r,a,n);case 3:e:{if(j9(t),e===null)throw Error(U(387));r=t.pendingProps,o=t.memoizedState,a=o.element,a9(e,t),Zc(t,r,null,n);var i=t.memoizedState;if(r=i.element,o.isDehydrated)if(o={element:r,isDehydrated:!1,cache:i.cache,pendingSuspenseBoundaries:i.pendingSuspenseBoundaries,transitions:i.transitions},t.updateQueue.baseState=o,t.memoizedState=o,t.flags&256){a=Ai(Error(U(423)),t),t=K1(e,t,r,n,a);break e}else if(r!==a){a=Ai(Error(U(424)),t),t=K1(e,t,r,n,a);break e}else for(Jt=va(t.stateNode.containerInfo.firstChild),tn=t,Be=!0,Mn=null,n=n9(t,null,r,n),t.child=n;n;)n.flags=n.flags&-3|4096,n=n.sibling;else{if(ki(),r===a){t=Ir(e,t,n);break e}jt(e,t,r,n)}t=t.child}return t;case 5:return o9(t),e===null&&wh(t),r=t.type,a=t.pendingProps,o=e!==null?e.memoizedProps:null,i=a.children,hh(r,a)?i=null:o!==null&&hh(r,o)&&(t.flags|=32),T9(e,t),jt(e,t,i,n),t.child;case 6:return e===null&&wh(t),null;case 13:return N9(e,t,n);case 4:return Jv(t,t.stateNode.containerInfo),r=t.pendingProps,e===null?t.child=Ci(t,null,r,n):jt(e,t,r,n),t.child;case 11:return r=t.type,a=t.pendingProps,a=t.elementType===r?a:_n(r,a),W1(e,t,r,a,n);case 7:return jt(e,t,t.pendingProps,n),t.child;case 8:return jt(e,t,t.pendingProps.children,n),t.child;case 12:return jt(e,t,t.pendingProps.children,n),t.child;case 10:e:{if(r=t.type._context,a=t.pendingProps,o=t.memoizedProps,i=a.value,je(Yc,r._currentValue),r._currentValue=i,o!==null)if(Gn(o.value,i)){if(o.children===a.children&&!Gt.current){t=Ir(e,t,n);break e}}else for(o=t.child,o!==null&&(o.return=t);o!==null;){var s=o.dependencies;if(s!==null){i=o.child;for(var l=s.firstContext;l!==null;){if(l.context===r){if(o.tag===1){l=Ar(-1,n&-n),l.tag=2;var u=o.updateQueue;if(u!==null){u=u.shared;var p=u.pending;p===null?l.next=l:(l.next=p.next,p.next=l),u.pending=l}}o.lanes|=n,l=o.alternate,l!==null&&(l.lanes|=n),bh(o.return,n,t),s.lanes|=n;break}l=l.next}}else if(o.tag===10)i=o.type===t.type?null:o.child;else if(o.tag===18){if(i=o.return,i===null)throw Error(U(341));i.lanes|=n,s=i.alternate,s!==null&&(s.lanes|=n),bh(i,n,t),i=o.sibling}else i=o.child;if(i!==null)i.return=o;else for(i=o;i!==null;){if(i===t){i=null;break}if(o=i.sibling,o!==null){o.return=i.return,i=o;break}i=i.return}o=i}jt(e,t,a.children,n),t=t.child}return t;case 9:return a=t.type,r=t.pendingProps.children,ai(t,n),a=gn(a),r=r(a),t.flags|=1,jt(e,t,r,n),t.child;case 14:return r=t.type,a=_n(r,t.pendingProps),a=_n(r.type,a),q1(e,t,r,a,n);case 15:return A9(e,t,t.type,t.pendingProps,n);case 17:return r=t.type,a=t.pendingProps,a=t.elementType===r?a:_n(r,a),_c(e,t),t.tag=1,Ut(r)?(e=!0,Vc(t)):e=!1,ai(t,n),k9(t,r,a),Sh(t,r,a,n),Ch(null,t,r,!0,e,n);case 19:return $9(e,t,n);case 22:return E9(e,t,n)}throw Error(U(156,t.tag))};function X9(e,t){return P5(e,t)}function AC(e,t,n,r){this.tag=e,this.key=n,this.sibling=this.child=this.return=this.stateNode=this.type=this.elementType=null,this.index=0,this.ref=null,this.pendingProps=t,this.dependencies=this.memoizedState=this.updateQueue=this.memoizedProps=null,this.mode=r,this.subtreeFlags=this.flags=0,this.deletions=null,this.childLanes=this.lanes=0,this.alternate=null}function dn(e,t,n,r){return new AC(e,t,n,r)}function my(e){return e=e.prototype,!(!e||!e.isReactComponent)}function EC(e){if(typeof e=="function")return my(e)?1:0;if(e!=null){if(e=e.$$typeof,e===$v)return 11;if(e===Mv)return 14}return 2}function wa(e,t){var n=e.alternate;return n===null?(n=dn(e.tag,t,e.key,e.mode),n.elementType=e.elementType,n.type=e.type,n.stateNode=e.stateNode,n.alternate=e,e.alternate=n):(n.pendingProps=t,n.type=e.type,n.flags=0,n.subtreeFlags=0,n.deletions=null),n.flags=e.flags&14680064,n.childLanes=e.childLanes,n.lanes=e.lanes,n.child=e.child,n.memoizedProps=e.memoizedProps,n.memoizedState=e.memoizedState,n.updateQueue=e.updateQueue,t=e.dependencies,n.dependencies=t===null?null:{lanes:t.lanes,firstContext:t.firstContext},n.sibling=e.sibling,n.index=e.index,n.ref=e.ref,n}function Tc(e,t,n,r,a,o){var i=2;if(r=e,typeof e=="function")my(e)&&(i=1);else if(typeof e=="string")i=5;else e:switch(e){case Bo:return so(n.children,a,o,t);case Nv:i=8,a|=8;break;case Vm:return e=dn(12,n,t,a|2),e.elementType=Vm,e.lanes=o,e;case Km:return e=dn(13,n,t,a),e.elementType=Km,e.lanes=o,e;case Xm:return e=dn(19,n,t,a),e.elementType=Xm,e.lanes=o,e;case o5:return Af(n,a,o,t);default:if(typeof e=="object"&&e!==null)switch(e.$$typeof){case r5:i=10;break e;case a5:i=9;break e;case $v:i=11;break e;case Mv:i=14;break e;case Zr:i=16,r=null;break e}throw Error(U(130,e==null?e:typeof e,""))}return t=dn(i,n,t,a),t.elementType=e,t.type=r,t.lanes=o,t}function so(e,t,n,r){return e=dn(7,e,r,t),e.lanes=n,e}function Af(e,t,n,r){return e=dn(22,e,r,t),e.elementType=o5,e.lanes=n,e.stateNode={isHidden:!1},e}function am(e,t,n){return e=dn(6,e,null,t),e.lanes=n,e}function om(e,t,n){return t=dn(4,e.children!==null?e.children:[],e.key,t),t.lanes=n,t.stateNode={containerInfo:e.containerInfo,pendingChildren:null,implementation:e.implementation},t}function TC(e,t,n,r,a){this.tag=t,this.containerInfo=e,this.finishedWork=this.pingCache=this.current=this.pendingChildren=null,this.timeoutHandle=-1,this.callbackNode=this.pendingContext=this.context=null,this.callbackPriority=0,this.eventTimes=Fd(0),this.expirationTimes=Fd(-1),this.entangledLanes=this.finishedLanes=this.mutableReadLanes=this.expiredLanes=this.pingedLanes=this.suspendedLanes=this.pendingLanes=0,this.entanglements=Fd(0),this.identifierPrefix=r,this.onRecoverableError=a,this.mutableSourceEagerHydrationData=null}function hy(e,t,n,r,a,o,i,s,l){return e=new TC(e,t,n,s,l),t===1?(t=1,o===!0&&(t|=8)):t=0,o=dn(3,null,null,t),e.current=o,o.stateNode=e,o.memoizedState={element:r,isDehydrated:n,cache:null,transitions:null,pendingSuspenseBoundaries:null},Zv(o),e}function jC(e,t,n){var r=3<arguments.length&&arguments[3]!==void 0?arguments[3]:null;return{$$typeof:Fo,key:r==null?null:""+r,children:e,containerInfo:t,implementation:n}}function Y9(e){if(!e)return Sa;e=e._reactInternals;e:{if(ko(e)!==e||e.tag!==1)throw Error(U(170));var t=e;do{switch(t.tag){case 3:t=t.stateNode.context;break e;case 1:if(Ut(t.type)){t=t.stateNode.__reactInternalMemoizedMergedChildContext;break e}}t=t.return}while(t!==null);throw Error(U(171))}if(e.tag===1){var n=e.type;if(Ut(n))return Y5(e,n,t)}return t}function Q9(e,t,n,r,a,o,i,s,l){return e=hy(n,r,!0,e,a,o,i,s,l),e.context=Y9(null),n=e.current,r=Rt(),a=xa(n),o=Ar(r,a),o.callback=t??null,ya(n,o,a),e.current.lanes=a,gu(e,a,r),Wt(e,r),e}function Ef(e,t,n,r){var a=t.current,o=Rt(),i=xa(a);return n=Y9(n),t.context===null?t.context=n:t.pendingContext=n,t=Ar(o,i),t.payload={element:e},r=r===void 0?null:r,r!==null&&(t.callback=r),e=ya(a,t,i),e!==null&&(Hn(e,a,i,o),Oc(e,a,i)),i}function ip(e){if(e=e.current,!e.child)return null;switch(e.child.tag){case 5:return e.child.stateNode;default:return e.child.stateNode}}function ax(e,t){if(e=e.memoizedState,e!==null&&e.dehydrated!==null){var n=e.retryLane;e.retryLane=n!==0&&n<t?n:t}}function vy(e,t){ax(e,t),(e=e.alternate)&&ax(e,t)}function NC(){return null}var Z9=typeof reportError=="function"?reportError:function(e){console.error(e)};function yy(e){this._internalRoot=e}Tf.prototype.render=yy.prototype.render=function(e){var t=this._internalRoot;if(t===null)throw Error(U(409));Ef(e,t,null,null)};Tf.prototype.unmount=yy.prototype.unmount=function(){var e=this._internalRoot;if(e!==null){this._internalRoot=null;var t=e.containerInfo;ho(function(){Ef(null,e,null,null)}),t[Mr]=null}};function Tf(e){this._internalRoot=e}Tf.prototype.unstable_scheduleHydration=function(e){if(e){var t=E5();e={blockedOn:null,target:e,priority:t};for(var n=0;n<ta.length&&t!==0&&t<ta[n].priority;n++);ta.splice(n,0,e),n===0&&j5(e)}};function gy(e){return!(!e||e.nodeType!==1&&e.nodeType!==9&&e.nodeType!==11)}function jf(e){return!(!e||e.nodeType!==1&&e.nodeType!==9&&e.nodeType!==11&&(e.nodeType!==8||e.nodeValue!==" react-mount-point-unstable "))}function ox(){}function $C(e,t,n,r,a){if(a){if(typeof r=="function"){var o=r;r=function(){var u=ip(i);o.call(u)}}var i=Q9(t,r,e,0,null,!1,!1,"",ox);return e._reactRootContainer=i,e[Mr]=i.current,xl(e.nodeType===8?e.parentNode:e),ho(),i}for(;a=e.lastChild;)e.removeChild(a);if(typeof r=="function"){var s=r;r=function(){var u=ip(l);s.call(u)}}var l=hy(e,0,!1,null,null,!1,!1,"",ox);return e._reactRootContainer=l,e[Mr]=l.current,xl(e.nodeType===8?e.parentNode:e),ho(function(){Ef(t,l,n,r)}),l}function Nf(e,t,n,r,a){var o=n._reactRootContainer;if(o){var i=o;if(typeof a=="function"){var s=a;a=function(){var l=ip(i);s.call(l)}}Ef(t,i,e,a)}else i=$C(n,t,e,a,r);return ip(i)}_5=function(e){switch(e.tag){case 3:var t=e.stateNode;if(t.current.memoizedState.isDehydrated){var n=Vs(t.pendingLanes);n!==0&&(Dv(t,n|1),Wt(t,Ke()),!(de&6)&&(Ei=Ke()+500,Ea()))}break;case 13:ho(function(){var r=Rr(e,1);if(r!==null){var a=Rt();Hn(r,e,1,a)}}),vy(e,1)}};Lv=function(e){if(e.tag===13){var t=Rr(e,134217728);if(t!==null){var n=Rt();Hn(t,e,134217728,n)}vy(e,134217728)}};A5=function(e){if(e.tag===13){var t=xa(e),n=Rr(e,t);if(n!==null){var r=Rt();Hn(n,e,t,r)}vy(e,t)}};E5=function(){return Oe};T5=function(e,t){var n=Oe;try{return Oe=e,t()}finally{Oe=n}};oh=function(e,t,n){switch(t){case"input":if(Zm(e,n),t=n.name,n.type==="radio"&&t!=null){for(n=e;n.parentNode;)n=n.parentNode;for(n=n.querySelectorAll("input[name="+JSON.stringify(""+t)+'][type="radio"]'),t=0;t<n.length;t++){var r=n[t];if(r!==e&&r.form===e.form){var a=Pf(r);if(!a)throw Error(U(90));s5(r),Zm(r,a)}}}break;case"textarea":u5(e,n);break;case"select":t=n.value,t!=null&&ei(e,!!n.multiple,t,!1)}};v5=py;y5=ho;var MC={usingClientEntryPoint:!1,Events:[wu,Uo,Pf,m5,h5,py]},Ns={findFiberByHostInstance:qa,bundleType:0,version:"18.3.1",rendererPackageName:"react-dom"},RC={bundleType:Ns.bundleType,version:Ns.version,rendererPackageName:Ns.rendererPackageName,rendererConfig:Ns.rendererConfig,overrideHookState:null,overrideHookStateDeletePath:null,overrideHookStateRenamePath:null,overrideProps:null,overridePropsDeletePath:null,overridePropsRenamePath:null,setErrorHandler:null,setSuspenseHandler:null,scheduleUpdate:null,currentDispatcherRef:Hr.ReactCurrentDispatcher,findHostInstanceByFiber:function(e){return e=w5(e),e===null?null:e.stateNode},findFiberByHostInstance:Ns.findFiberByHostInstance||NC,findHostInstancesForRefresh:null,scheduleRefresh:null,scheduleRoot:null,setRefreshHandler:null,getCurrentFiber:null,reconcilerVersion:"18.3.1-next-f1338f8080-20240426"};if(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__<"u"){var Qu=__REACT_DEVTOOLS_GLOBAL_HOOK__;if(!Qu.isDisabled&&Qu.supportsFiber)try{gf=Qu.inject(RC),sr=Qu}catch{}}an.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED=MC;an.createPortal=function(e,t){var n=2<arguments.length&&arguments[2]!==void 0?arguments[2]:null;if(!gy(t))throw Error(U(200));return jC(e,t,null,n)};an.createRoot=function(e,t){if(!gy(e))throw Error(U(299));var n=!1,r="",a=Z9;return t!=null&&(t.unstable_strictMode===!0&&(n=!0),t.identifierPrefix!==void 0&&(r=t.identifierPrefix),t.onRecoverableError!==void 0&&(a=t.onRecoverableError)),t=hy(e,1,!1,null,null,n,!1,r,a),e[Mr]=t.current,xl(e.nodeType===8?e.parentNode:e),new yy(t)};an.findDOMNode=function(e){if(e==null)return null;if(e.nodeType===1)return e;var t=e._reactInternals;if(t===void 0)throw typeof e.render=="function"?Error(U(188)):(e=Object.keys(e).join(","),Error(U(268,e)));return e=w5(t),e=e===null?null:e.stateNode,e};an.flushSync=function(e){return ho(e)};an.hydrate=function(e,t,n){if(!jf(t))throw Error(U(200));return Nf(null,e,t,!0,n)};an.hydrateRoot=function(e,t,n){if(!gy(e))throw Error(U(405));var r=n!=null&&n.hydratedSources||null,a=!1,o="",i=Z9;if(n!=null&&(n.unstable_strictMode===!0&&(a=!0),n.identifierPrefix!==void 0&&(o=n.identifierPrefix),n.onRecoverableError!==void 0&&(i=n.onRecoverableError)),t=Q9(t,null,e,1,n??null,a,!1,o,i),e[Mr]=t.current,xl(e),r)for(e=0;e<r.length;e++)n=r[e],a=n._getVersion,a=a(n._source),t.mutableSourceEagerHydrationData==null?t.mutableSourceEagerHydrationData=[n,a]:t.mutableSourceEagerHydrationData.push(n,a);return new Tf(t)};an.render=function(e,t,n){if(!jf(t))throw Error(U(200));return Nf(null,e,t,!1,n)};an.unmountComponentAtNode=function(e){if(!jf(e))throw Error(U(40));return e._reactRootContainer?(ho(function(){Nf(null,null,e,!1,function(){e._reactRootContainer=null,e[Mr]=null})}),!0):!1};an.unstable_batchedUpdates=py;an.unstable_renderSubtreeIntoContainer=function(e,t,n,r){if(!jf(n))throw Error(U(200));if(e==null||e._reactInternals===void 0)throw Error(U(38));return Nf(e,t,n,!1,r)};an.version="18.3.1-next-f1338f8080-20240426";function J9(){if(!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__>"u"||typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE!="function"))try{__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(J9)}catch(e){console.error(e)}}J9(),J2.exports=an;var os=J2.exports;const IC=_e(os);var ix=os;Ic.createRoot=ix.createRoot,Ic.hydrateRoot=ix.hydrateRoot;function sx(e,[t,n]){return Math.min(n,Math.max(t,e))}function fe(e,t,{checkForDefaultPrevented:n=!0}={}){return function(a){if(e==null||e(a),n===!1||!a.defaultPrevented)return t==null?void 0:t(a)}}function is(e,t=[]){let n=[];function r(o,i){const s=k.createContext(i),l=n.length;n=[...n,i];const u=c=>{var v;const{scope:f,children:m,...h}=c,d=((v=f==null?void 0:f[e])==null?void 0:v[l])||s,g=k.useMemo(()=>h,Object.values(h));return b.jsx(d.Provider,{value:g,children:m})};u.displayName=o+"Provider";function p(c,f){var d;const m=((d=f==null?void 0:f[e])==null?void 0:d[l])||s,h=k.useContext(m);if(h)return h;if(i!==void 0)return i;throw new Error(`\`${c}\` must be used within \`${o}\``)}return[u,p]}const a=()=>{const o=n.map(i=>k.createContext(i));return function(s){const l=(s==null?void 0:s[e])||o;return k.useMemo(()=>({[`__scope${e}`]:{...s,[e]:l}}),[s,l])}};return a.scopeName=e,[r,DC(a,...t)]}function DC(...e){const t=e[0];if(e.length===1)return t;const n=()=>{const r=e.map(a=>({useScope:a(),scopeName:a.scopeName}));return function(o){const i=r.reduce((s,{useScope:l,scopeName:u})=>{const c=l(o)[`__scope${u}`];return{...s,...c}},{});return k.useMemo(()=>({[`__scope${t.scopeName}`]:i}),[i])}};return n.scopeName=t.scopeName,n}function lx(e,t){if(typeof e=="function")return e(t);e!=null&&(e.current=t)}function e6(...e){return t=>{let n=!1;const r=e.map(a=>{const o=lx(a,t);return!n&&typeof o=="function"&&(n=!0),o});if(n)return()=>{for(let a=0;a<r.length;a++){const o=r[a];typeof o=="function"?o():lx(e[a],null)}}}}function Xe(...e){return k.useCallback(e6(...e),e)}function Al(e){const t=FC(e),n=k.forwardRef((r,a)=>{const{children:o,...i}=r,s=k.Children.toArray(o),l=s.find(zC);if(l){const u=l.props.children,p=s.map(c=>c===l?k.Children.count(u)>1?k.Children.only(null):k.isValidElement(u)?u.props.children:null:c);return b.jsx(t,{...i,ref:a,children:k.isValidElement(u)?k.cloneElement(u,void 0,p):null})}return b.jsx(t,{...i,ref:a,children:o})});return n.displayName=`${e}.Slot`,n}var LC=Al("Slot");function FC(e){const t=k.forwardRef((n,r)=>{const{children:a,...o}=n;if(k.isValidElement(a)){const i=GC(a),s=HC(o,a.props);return a.type!==k.Fragment&&(s.ref=r?e6(r,i):i),k.cloneElement(a,s)}return k.Children.count(a)>1?k.Children.only(null):null});return t.displayName=`${e}.SlotClone`,t}var t6=Symbol("radix.slottable");function BC(e){const t=({children:n})=>b.jsx(b.Fragment,{children:n});return t.displayName=`${e}.Slottable`,t.__radixId=t6,t}function zC(e){return k.isValidElement(e)&&typeof e.type=="function"&&"__radixId"in e.type&&e.type.__radixId===t6}function HC(e,t){const n={...t};for(const r in t){const a=e[r],o=t[r];/^on[A-Z]/.test(r)?a&&o?n[r]=(...s)=>{const l=o(...s);return a(...s),l}:a&&(n[r]=a):r==="style"?n[r]={...a,...o}:r==="className"&&(n[r]=[a,o].filter(Boolean).join(" "))}return{...e,...n}}function GC(e){var r,a;let t=(r=Object.getOwnPropertyDescriptor(e.props,"ref"))==null?void 0:r.get,n=t&&"isReactWarning"in t&&t.isReactWarning;return n?e.ref:(t=(a=Object.getOwnPropertyDescriptor(e,"ref"))==null?void 0:a.get,n=t&&"isReactWarning"in t&&t.isReactWarning,n?e.props.ref:e.props.ref||e.ref)}function n6(e){const t=e+"CollectionProvider",[n,r]=is(t),[a,o]=n(t,{collectionRef:{current:null},itemMap:new Map}),i=d=>{const{scope:g,children:v}=d,y=E.useRef(null),x=E.useRef(new Map).current;return b.jsx(a,{scope:g,itemMap:x,collectionRef:y,children:v})};i.displayName=t;const s=e+"CollectionSlot",l=Al(s),u=E.forwardRef((d,g)=>{const{scope:v,children:y}=d,x=o(s,v),P=Xe(g,x.collectionRef);return b.jsx(l,{ref:P,children:y})});u.displayName=s;const p=e+"CollectionItemSlot",c="data-radix-collection-item",f=Al(p),m=E.forwardRef((d,g)=>{const{scope:v,children:y,...x}=d,P=E.useRef(null),w=Xe(g,P),S=o(p,v);return E.useEffect(()=>(S.itemMap.set(P,{ref:P,...x}),()=>void S.itemMap.delete(P))),b.jsx(f,{[c]:"",ref:w,children:y})});m.displayName=p;function h(d){const g=o(e+"CollectionConsumer",d);return E.useCallback(()=>{const y=g.collectionRef.current;if(!y)return[];const x=Array.from(y.querySelectorAll(`[${c}]`));return Array.from(g.itemMap.values()).sort((S,O)=>x.indexOf(S.ref.current)-x.indexOf(O.ref.current))},[g.collectionRef,g.itemMap])}return[{Provider:i,Slot:u,ItemSlot:m},h,r]}var UC=k.createContext(void 0);function xy(e){const t=k.useContext(UC);return e||t||"ltr"}var WC=["a","button","div","form","h2","h3","img","input","label","li","nav","ol","p","select","span","svg","ul"],Ae=WC.reduce((e,t)=>{const n=Al(`Primitive.${t}`),r=k.forwardRef((a,o)=>{const{asChild:i,...s}=a,l=i?n:t;return typeof window<"u"&&(window[Symbol.for("radix-ui")]=!0),b.jsx(l,{...s,ref:o})});return r.displayName=`Primitive.${t}`,{...e,[t]:r}},{});function qC(e,t){e&&os.flushSync(()=>e.dispatchEvent(t))}function Oa(e){const t=k.useRef(e);return k.useEffect(()=>{t.current=e}),k.useMemo(()=>(...n)=>{var r;return(r=t.current)==null?void 0:r.call(t,...n)},[])}function VC(e,t=globalThis==null?void 0:globalThis.document){const n=Oa(e);k.useEffect(()=>{const r=a=>{a.key==="Escape"&&n(a)};return t.addEventListener("keydown",r,{capture:!0}),()=>t.removeEventListener("keydown",r,{capture:!0})},[n,t])}var KC="DismissableLayer",Lh="dismissableLayer.update",XC="dismissableLayer.pointerDownOutside",YC="dismissableLayer.focusOutside",ux,r6=k.createContext({layers:new Set,layersWithOutsidePointerEventsDisabled:new Set,branches:new Set}),wy=k.forwardRef((e,t)=>{const{disableOutsidePointerEvents:n=!1,onEscapeKeyDown:r,onPointerDownOutside:a,onFocusOutside:o,onInteractOutside:i,onDismiss:s,...l}=e,u=k.useContext(r6),[p,c]=k.useState(null),f=(p==null?void 0:p.ownerDocument)??(globalThis==null?void 0:globalThis.document),[,m]=k.useState({}),h=Xe(t,O=>c(O)),d=Array.from(u.layers),[g]=[...u.layersWithOutsidePointerEventsDisabled].slice(-1),v=d.indexOf(g),y=p?d.indexOf(p):-1,x=u.layersWithOutsidePointerEventsDisabled.size>0,P=y>=v,w=JC(O=>{const C=O.target,_=[...u.branches].some(T=>T.contains(C));!P||_||(a==null||a(O),i==null||i(O),O.defaultPrevented||s==null||s())},f),S=e_(O=>{const C=O.target;[...u.branches].some(T=>T.contains(C))||(o==null||o(O),i==null||i(O),O.defaultPrevented||s==null||s())},f);return VC(O=>{y===u.layers.size-1&&(r==null||r(O),!O.defaultPrevented&&s&&(O.preventDefault(),s()))},f),k.useEffect(()=>{if(p)return n&&(u.layersWithOutsidePointerEventsDisabled.size===0&&(ux=f.body.style.pointerEvents,f.body.style.pointerEvents="none"),u.layersWithOutsidePointerEventsDisabled.add(p)),u.layers.add(p),cx(),()=>{n&&u.layersWithOutsidePointerEventsDisabled.size===1&&(f.body.style.pointerEvents=ux)}},[p,f,n,u]),k.useEffect(()=>()=>{p&&(u.layers.delete(p),u.layersWithOutsidePointerEventsDisabled.delete(p),cx())},[p,u]),k.useEffect(()=>{const O=()=>m({});return document.addEventListener(Lh,O),()=>document.removeEventListener(Lh,O)},[]),b.jsx(Ae.div,{...l,ref:h,style:{pointerEvents:x?P?"auto":"none":void 0,...e.style},onFocusCapture:fe(e.onFocusCapture,S.onFocusCapture),onBlurCapture:fe(e.onBlurCapture,S.onBlurCapture),onPointerDownCapture:fe(e.onPointerDownCapture,w.onPointerDownCapture)})});wy.displayName=KC;var QC="DismissableLayerBranch",ZC=k.forwardRef((e,t)=>{const n=k.useContext(r6),r=k.useRef(null),a=Xe(t,r);return k.useEffect(()=>{const o=r.current;if(o)return n.branches.add(o),()=>{n.branches.delete(o)}},[n.branches]),b.jsx(Ae.div,{...e,ref:a})});ZC.displayName=QC;function JC(e,t=globalThis==null?void 0:globalThis.document){const n=Oa(e),r=k.useRef(!1),a=k.useRef(()=>{});return k.useEffect(()=>{const o=s=>{if(s.target&&!r.current){let l=function(){a6(XC,n,u,{discrete:!0})};const u={originalEvent:s};s.pointerType==="touch"?(t.removeEventListener("click",a.current),a.current=l,t.addEventListener("click",a.current,{once:!0})):l()}else t.removeEventListener("click",a.current);r.current=!1},i=window.setTimeout(()=>{t.addEventListener("pointerdown",o)},0);return()=>{window.clearTimeout(i),t.removeEventListener("pointerdown",o),t.removeEventListener("click",a.current)}},[t,n]),{onPointerDownCapture:()=>r.current=!0}}function e_(e,t=globalThis==null?void 0:globalThis.document){const n=Oa(e),r=k.useRef(!1);return k.useEffect(()=>{const a=o=>{o.target&&!r.current&&a6(YC,n,{originalEvent:o},{discrete:!1})};return t.addEventListener("focusin",a),()=>t.removeEventListener("focusin",a)},[t,n]),{onFocusCapture:()=>r.current=!0,onBlurCapture:()=>r.current=!1}}function cx(){const e=new CustomEvent(Lh);document.dispatchEvent(e)}function a6(e,t,n,{discrete:r}){const a=n.originalEvent.target,o=new CustomEvent(e,{bubbles:!1,cancelable:!0,detail:n});t&&a.addEventListener(e,t,{once:!0}),r?qC(a,o):a.dispatchEvent(o)}var im=0;function t_(){k.useEffect(()=>{const e=document.querySelectorAll("[data-radix-focus-guard]");return document.body.insertAdjacentElement("afterbegin",e[0]??px()),document.body.insertAdjacentElement("beforeend",e[1]??px()),im++,()=>{im===1&&document.querySelectorAll("[data-radix-focus-guard]").forEach(t=>t.remove()),im--}},[])}function px(){const e=document.createElement("span");return e.setAttribute("data-radix-focus-guard",""),e.tabIndex=0,e.style.outline="none",e.style.opacity="0",e.style.position="fixed",e.style.pointerEvents="none",e}var sm="focusScope.autoFocusOnMount",lm="focusScope.autoFocusOnUnmount",fx={bubbles:!1,cancelable:!0},n_="FocusScope",o6=k.forwardRef((e,t)=>{const{loop:n=!1,trapped:r=!1,onMountAutoFocus:a,onUnmountAutoFocus:o,...i}=e,[s,l]=k.useState(null),u=Oa(a),p=Oa(o),c=k.useRef(null),f=Xe(t,d=>l(d)),m=k.useRef({paused:!1,pause(){this.paused=!0},resume(){this.paused=!1}}).current;k.useEffect(()=>{if(r){let d=function(x){if(m.paused||!s)return;const P=x.target;s.contains(P)?c.current=P:Xr(c.current,{select:!0})},g=function(x){if(m.paused||!s)return;const P=x.relatedTarget;P!==null&&(s.contains(P)||Xr(c.current,{select:!0}))},v=function(x){if(document.activeElement===document.body)for(const w of x)w.removedNodes.length>0&&Xr(s)};document.addEventListener("focusin",d),document.addEventListener("focusout",g);const y=new MutationObserver(v);return s&&y.observe(s,{childList:!0,subtree:!0}),()=>{document.removeEventListener("focusin",d),document.removeEventListener("focusout",g),y.disconnect()}}},[r,s,m.paused]),k.useEffect(()=>{if(s){mx.add(m);const d=document.activeElement;if(!s.contains(d)){const v=new CustomEvent(sm,fx);s.addEventListener(sm,u),s.dispatchEvent(v),v.defaultPrevented||(r_(l_(i6(s)),{select:!0}),document.activeElement===d&&Xr(s))}return()=>{s.removeEventListener(sm,u),setTimeout(()=>{const v=new CustomEvent(lm,fx);s.addEventListener(lm,p),s.dispatchEvent(v),v.defaultPrevented||Xr(d??document.body,{select:!0}),s.removeEventListener(lm,p),mx.remove(m)},0)}}},[s,u,p,m]);const h=k.useCallback(d=>{if(!n&&!r||m.paused)return;const g=d.key==="Tab"&&!d.altKey&&!d.ctrlKey&&!d.metaKey,v=document.activeElement;if(g&&v){const y=d.currentTarget,[x,P]=a_(y);x&&P?!d.shiftKey&&v===P?(d.preventDefault(),n&&Xr(x,{select:!0})):d.shiftKey&&v===x&&(d.preventDefault(),n&&Xr(P,{select:!0})):v===y&&d.preventDefault()}},[n,r,m.paused]);return b.jsx(Ae.div,{tabIndex:-1,...i,ref:f,onKeyDown:h})});o6.displayName=n_;function r_(e,{select:t=!1}={}){const n=document.activeElement;for(const r of e)if(Xr(r,{select:t}),document.activeElement!==n)return}function a_(e){const t=i6(e),n=dx(t,e),r=dx(t.reverse(),e);return[n,r]}function i6(e){const t=[],n=document.createTreeWalker(e,NodeFilter.SHOW_ELEMENT,{acceptNode:r=>{const a=r.tagName==="INPUT"&&r.type==="hidden";return r.disabled||r.hidden||a?NodeFilter.FILTER_SKIP:r.tabIndex>=0?NodeFilter.FILTER_ACCEPT:NodeFilter.FILTER_SKIP}});for(;n.nextNode();)t.push(n.currentNode);return t}function dx(e,t){for(const n of e)if(!o_(n,{upTo:t}))return n}function o_(e,{upTo:t}){if(getComputedStyle(e).visibility==="hidden")return!0;for(;e;){if(t!==void 0&&e===t)return!1;if(getComputedStyle(e).display==="none")return!0;e=e.parentElement}return!1}function i_(e){return e instanceof HTMLInputElement&&"select"in e}function Xr(e,{select:t=!1}={}){if(e&&e.focus){const n=document.activeElement;e.focus({preventScroll:!0}),e!==n&&i_(e)&&t&&e.select()}}var mx=s_();function s_(){let e=[];return{add(t){const n=e[0];t!==n&&(n==null||n.pause()),e=hx(e,t),e.unshift(t)},remove(t){var n;e=hx(e,t),(n=e[0])==null||n.resume()}}}function hx(e,t){const n=[...e],r=n.indexOf(t);return r!==-1&&n.splice(r,1),n}function l_(e){return e.filter(t=>t.tagName!=="A")}var Et=globalThis!=null&&globalThis.document?k.useLayoutEffect:()=>{},u_=Q2[" useId ".trim().toString()]||(()=>{}),c_=0;function Pu(e){const[t,n]=k.useState(u_());return Et(()=>{n(r=>r??String(c_++))},[e]),t?`radix-${t}`:""}const p_=["top","right","bottom","left"],ka=Math.min,Qt=Math.max,sp=Math.round,Zu=Math.floor,ur=e=>({x:e,y:e}),f_={left:"right",right:"left",bottom:"top",top:"bottom"},d_={start:"end",end:"start"};function Fh(e,t,n){return Qt(e,ka(t,n))}function Dr(e,t){return typeof e=="function"?e(t):e}function Lr(e){return e.split("-")[0]}function ss(e){return e.split("-")[1]}function by(e){return e==="x"?"y":"x"}function Py(e){return e==="y"?"height":"width"}function nr(e){return["top","bottom"].includes(Lr(e))?"y":"x"}function Sy(e){return by(nr(e))}function m_(e,t,n){n===void 0&&(n=!1);const r=ss(e),a=Sy(e),o=Py(a);let i=a==="x"?r===(n?"end":"start")?"right":"left":r==="start"?"bottom":"top";return t.reference[o]>t.floating[o]&&(i=lp(i)),[i,lp(i)]}function h_(e){const t=lp(e);return[Bh(e),t,Bh(t)]}function Bh(e){return e.replace(/start|end/g,t=>d_[t])}function v_(e,t,n){const r=["left","right"],a=["right","left"],o=["top","bottom"],i=["bottom","top"];switch(e){case"top":case"bottom":return n?t?a:r:t?r:a;case"left":case"right":return t?o:i;default:return[]}}function y_(e,t,n,r){const a=ss(e);let o=v_(Lr(e),n==="start",r);return a&&(o=o.map(i=>i+"-"+a),t&&(o=o.concat(o.map(Bh)))),o}function lp(e){return e.replace(/left|right|bottom|top/g,t=>f_[t])}function g_(e){return{top:0,right:0,bottom:0,left:0,...e}}function s6(e){return typeof e!="number"?g_(e):{top:e,right:e,bottom:e,left:e}}function up(e){const{x:t,y:n,width:r,height:a}=e;return{width:r,height:a,top:n,left:t,right:t+r,bottom:n+a,x:t,y:n}}function vx(e,t,n){let{reference:r,floating:a}=e;const o=nr(t),i=Sy(t),s=Py(i),l=Lr(t),u=o==="y",p=r.x+r.width/2-a.width/2,c=r.y+r.height/2-a.height/2,f=r[s]/2-a[s]/2;let m;switch(l){case"top":m={x:p,y:r.y-a.height};break;case"bottom":m={x:p,y:r.y+r.height};break;case"right":m={x:r.x+r.width,y:c};break;case"left":m={x:r.x-a.width,y:c};break;default:m={x:r.x,y:r.y}}switch(ss(t)){case"start":m[i]-=f*(n&&u?-1:1);break;case"end":m[i]+=f*(n&&u?-1:1);break}return m}const x_=async(e,t,n)=>{const{placement:r="bottom",strategy:a="absolute",middleware:o=[],platform:i}=n,s=o.filter(Boolean),l=await(i.isRTL==null?void 0:i.isRTL(t));let u=await i.getElementRects({reference:e,floating:t,strategy:a}),{x:p,y:c}=vx(u,r,l),f=r,m={},h=0;for(let d=0;d<s.length;d++){const{name:g,fn:v}=s[d],{x:y,y:x,data:P,reset:w}=await v({x:p,y:c,initialPlacement:r,placement:f,strategy:a,middlewareData:m,rects:u,platform:i,elements:{reference:e,floating:t}});p=y??p,c=x??c,m={...m,[g]:{...m[g],...P}},w&&h<=50&&(h++,typeof w=="object"&&(w.placement&&(f=w.placement),w.rects&&(u=w.rects===!0?await i.getElementRects({reference:e,floating:t,strategy:a}):w.rects),{x:p,y:c}=vx(u,f,l)),d=-1)}return{x:p,y:c,placement:f,strategy:a,middlewareData:m}};async function El(e,t){var n;t===void 0&&(t={});const{x:r,y:a,platform:o,rects:i,elements:s,strategy:l}=e,{boundary:u="clippingAncestors",rootBoundary:p="viewport",elementContext:c="floating",altBoundary:f=!1,padding:m=0}=Dr(t,e),h=s6(m),g=s[f?c==="floating"?"reference":"floating":c],v=up(await o.getClippingRect({element:(n=await(o.isElement==null?void 0:o.isElement(g)))==null||n?g:g.contextElement||await(o.getDocumentElement==null?void 0:o.getDocumentElement(s.floating)),boundary:u,rootBoundary:p,strategy:l})),y=c==="floating"?{x:r,y:a,width:i.floating.width,height:i.floating.height}:i.reference,x=await(o.getOffsetParent==null?void 0:o.getOffsetParent(s.floating)),P=await(o.isElement==null?void 0:o.isElement(x))?await(o.getScale==null?void 0:o.getScale(x))||{x:1,y:1}:{x:1,y:1},w=up(o.convertOffsetParentRelativeRectToViewportRelativeRect?await o.convertOffsetParentRelativeRectToViewportRelativeRect({elements:s,rect:y,offsetParent:x,strategy:l}):y);return{top:(v.top-w.top+h.top)/P.y,bottom:(w.bottom-v.bottom+h.bottom)/P.y,left:(v.left-w.left+h.left)/P.x,right:(w.right-v.right+h.right)/P.x}}const w_=e=>({name:"arrow",options:e,async fn(t){const{x:n,y:r,placement:a,rects:o,platform:i,elements:s,middlewareData:l}=t,{element:u,padding:p=0}=Dr(e,t)||{};if(u==null)return{};const c=s6(p),f={x:n,y:r},m=Sy(a),h=Py(m),d=await i.getDimensions(u),g=m==="y",v=g?"top":"left",y=g?"bottom":"right",x=g?"clientHeight":"clientWidth",P=o.reference[h]+o.reference[m]-f[m]-o.floating[h],w=f[m]-o.reference[m],S=await(i.getOffsetParent==null?void 0:i.getOffsetParent(u));let O=S?S[x]:0;(!O||!await(i.isElement==null?void 0:i.isElement(S)))&&(O=s.floating[x]||o.floating[h]);const C=P/2-w/2,_=O/2-d[h]/2-1,T=ka(c[v],_),A=ka(c[y],_),j=T,$=O-d[h]-A,M=O/2-d[h]/2+C,I=Fh(j,M,$),R=!l.arrow&&ss(a)!=null&&M!==I&&o.reference[h]/2-(M<j?T:A)-d[h]/2<0,L=R?M<j?M-j:M-$:0;return{[m]:f[m]+L,data:{[m]:I,centerOffset:M-I-L,...R&&{alignmentOffset:L}},reset:R}}}),b_=function(e){return e===void 0&&(e={}),{name:"flip",options:e,async fn(t){var n,r;const{placement:a,middlewareData:o,rects:i,initialPlacement:s,platform:l,elements:u}=t,{mainAxis:p=!0,crossAxis:c=!0,fallbackPlacements:f,fallbackStrategy:m="bestFit",fallbackAxisSideDirection:h="none",flipAlignment:d=!0,...g}=Dr(e,t);if((n=o.arrow)!=null&&n.alignmentOffset)return{};const v=Lr(a),y=nr(s),x=Lr(s)===s,P=await(l.isRTL==null?void 0:l.isRTL(u.floating)),w=f||(x||!d?[lp(s)]:h_(s)),S=h!=="none";!f&&S&&w.push(...y_(s,d,h,P));const O=[s,...w],C=await El(t,g),_=[];let T=((r=o.flip)==null?void 0:r.overflows)||[];if(p&&_.push(C[v]),c){const M=m_(a,i,P);_.push(C[M[0]],C[M[1]])}if(T=[...T,{placement:a,overflows:_}],!_.every(M=>M<=0)){var A,j;const M=(((A=o.flip)==null?void 0:A.index)||0)+1,I=O[M];if(I&&(!(c==="alignment"?y!==nr(I):!1)||T.every(N=>N.overflows[0]>0&&nr(N.placement)===y)))return{data:{index:M,overflows:T},reset:{placement:I}};let R=(j=T.filter(L=>L.overflows[0]<=0).sort((L,N)=>L.overflows[1]-N.overflows[1])[0])==null?void 0:j.placement;if(!R)switch(m){case"bestFit":{var $;const L=($=T.filter(N=>{if(S){const D=nr(N.placement);return D===y||D==="y"}return!0}).map(N=>[N.placement,N.overflows.filter(D=>D>0).reduce((D,H)=>D+H,0)]).sort((N,D)=>N[1]-D[1])[0])==null?void 0:$[0];L&&(R=L);break}case"initialPlacement":R=s;break}if(a!==R)return{reset:{placement:R}}}return{}}}};function yx(e,t){return{top:e.top-t.height,right:e.right-t.width,bottom:e.bottom-t.height,left:e.left-t.width}}function gx(e){return p_.some(t=>e[t]>=0)}const P_=function(e){return e===void 0&&(e={}),{name:"hide",options:e,async fn(t){const{rects:n}=t,{strategy:r="referenceHidden",...a}=Dr(e,t);switch(r){case"referenceHidden":{const o=await El(t,{...a,elementContext:"reference"}),i=yx(o,n.reference);return{data:{referenceHiddenOffsets:i,referenceHidden:gx(i)}}}case"escaped":{const o=await El(t,{...a,altBoundary:!0}),i=yx(o,n.floating);return{data:{escapedOffsets:i,escaped:gx(i)}}}default:return{}}}}};async function S_(e,t){const{placement:n,platform:r,elements:a}=e,o=await(r.isRTL==null?void 0:r.isRTL(a.floating)),i=Lr(n),s=ss(n),l=nr(n)==="y",u=["left","top"].includes(i)?-1:1,p=o&&l?-1:1,c=Dr(t,e);let{mainAxis:f,crossAxis:m,alignmentAxis:h}=typeof c=="number"?{mainAxis:c,crossAxis:0,alignmentAxis:null}:{mainAxis:c.mainAxis||0,crossAxis:c.crossAxis||0,alignmentAxis:c.alignmentAxis};return s&&typeof h=="number"&&(m=s==="end"?h*-1:h),l?{x:m*p,y:f*u}:{x:f*u,y:m*p}}const O_=function(e){return e===void 0&&(e=0),{name:"offset",options:e,async fn(t){var n,r;const{x:a,y:o,placement:i,middlewareData:s}=t,l=await S_(t,e);return i===((n=s.offset)==null?void 0:n.placement)&&(r=s.arrow)!=null&&r.alignmentOffset?{}:{x:a+l.x,y:o+l.y,data:{...l,placement:i}}}}},k_=function(e){return e===void 0&&(e={}),{name:"shift",options:e,async fn(t){const{x:n,y:r,placement:a}=t,{mainAxis:o=!0,crossAxis:i=!1,limiter:s={fn:g=>{let{x:v,y}=g;return{x:v,y}}},...l}=Dr(e,t),u={x:n,y:r},p=await El(t,l),c=nr(Lr(a)),f=by(c);let m=u[f],h=u[c];if(o){const g=f==="y"?"top":"left",v=f==="y"?"bottom":"right",y=m+p[g],x=m-p[v];m=Fh(y,m,x)}if(i){const g=c==="y"?"top":"left",v=c==="y"?"bottom":"right",y=h+p[g],x=h-p[v];h=Fh(y,h,x)}const d=s.fn({...t,[f]:m,[c]:h});return{...d,data:{x:d.x-n,y:d.y-r,enabled:{[f]:o,[c]:i}}}}}},C_=function(e){return e===void 0&&(e={}),{options:e,fn(t){const{x:n,y:r,placement:a,rects:o,middlewareData:i}=t,{offset:s=0,mainAxis:l=!0,crossAxis:u=!0}=Dr(e,t),p={x:n,y:r},c=nr(a),f=by(c);let m=p[f],h=p[c];const d=Dr(s,t),g=typeof d=="number"?{mainAxis:d,crossAxis:0}:{mainAxis:0,crossAxis:0,...d};if(l){const x=f==="y"?"height":"width",P=o.reference[f]-o.floating[x]+g.mainAxis,w=o.reference[f]+o.reference[x]-g.mainAxis;m<P?m=P:m>w&&(m=w)}if(u){var v,y;const x=f==="y"?"width":"height",P=["top","left"].includes(Lr(a)),w=o.reference[c]-o.floating[x]+(P&&((v=i.offset)==null?void 0:v[c])||0)+(P?0:g.crossAxis),S=o.reference[c]+o.reference[x]+(P?0:((y=i.offset)==null?void 0:y[c])||0)-(P?g.crossAxis:0);h<w?h=w:h>S&&(h=S)}return{[f]:m,[c]:h}}}},__=function(e){return e===void 0&&(e={}),{name:"size",options:e,async fn(t){var n,r;const{placement:a,rects:o,platform:i,elements:s}=t,{apply:l=()=>{},...u}=Dr(e,t),p=await El(t,u),c=Lr(a),f=ss(a),m=nr(a)==="y",{width:h,height:d}=o.floating;let g,v;c==="top"||c==="bottom"?(g=c,v=f===(await(i.isRTL==null?void 0:i.isRTL(s.floating))?"start":"end")?"left":"right"):(v=c,g=f==="end"?"top":"bottom");const y=d-p.top-p.bottom,x=h-p.left-p.right,P=ka(d-p[g],y),w=ka(h-p[v],x),S=!t.middlewareData.shift;let O=P,C=w;if((n=t.middlewareData.shift)!=null&&n.enabled.x&&(C=x),(r=t.middlewareData.shift)!=null&&r.enabled.y&&(O=y),S&&!f){const T=Qt(p.left,0),A=Qt(p.right,0),j=Qt(p.top,0),$=Qt(p.bottom,0);m?C=h-2*(T!==0||A!==0?T+A:Qt(p.left,p.right)):O=d-2*(j!==0||$!==0?j+$:Qt(p.top,p.bottom))}await l({...t,availableWidth:C,availableHeight:O});const _=await i.getDimensions(s.floating);return h!==_.width||d!==_.height?{reset:{rects:!0}}:{}}}};function $f(){return typeof window<"u"}function ls(e){return l6(e)?(e.nodeName||"").toLowerCase():"#document"}function nn(e){var t;return(e==null||(t=e.ownerDocument)==null?void 0:t.defaultView)||window}function hr(e){var t;return(t=(l6(e)?e.ownerDocument:e.document)||window.document)==null?void 0:t.documentElement}function l6(e){return $f()?e instanceof Node||e instanceof nn(e).Node:!1}function Un(e){return $f()?e instanceof Element||e instanceof nn(e).Element:!1}function fr(e){return $f()?e instanceof HTMLElement||e instanceof nn(e).HTMLElement:!1}function xx(e){return!$f()||typeof ShadowRoot>"u"?!1:e instanceof ShadowRoot||e instanceof nn(e).ShadowRoot}function Su(e){const{overflow:t,overflowX:n,overflowY:r,display:a}=Wn(e);return/auto|scroll|overlay|hidden|clip/.test(t+r+n)&&!["inline","contents"].includes(a)}function A_(e){return["table","td","th"].includes(ls(e))}function Mf(e){return[":popover-open",":modal"].some(t=>{try{return e.matches(t)}catch{return!1}})}function Oy(e){const t=ky(),n=Un(e)?Wn(e):e;return["transform","translate","scale","rotate","perspective"].some(r=>n[r]?n[r]!=="none":!1)||(n.containerType?n.containerType!=="normal":!1)||!t&&(n.backdropFilter?n.backdropFilter!=="none":!1)||!t&&(n.filter?n.filter!=="none":!1)||["transform","translate","scale","rotate","perspective","filter"].some(r=>(n.willChange||"").includes(r))||["paint","layout","strict","content"].some(r=>(n.contain||"").includes(r))}function E_(e){let t=Ca(e);for(;fr(t)&&!Ti(t);){if(Oy(t))return t;if(Mf(t))return null;t=Ca(t)}return null}function ky(){return typeof CSS>"u"||!CSS.supports?!1:CSS.supports("-webkit-backdrop-filter","none")}function Ti(e){return["html","body","#document"].includes(ls(e))}function Wn(e){return nn(e).getComputedStyle(e)}function Rf(e){return Un(e)?{scrollLeft:e.scrollLeft,scrollTop:e.scrollTop}:{scrollLeft:e.scrollX,scrollTop:e.scrollY}}function Ca(e){if(ls(e)==="html")return e;const t=e.assignedSlot||e.parentNode||xx(e)&&e.host||hr(e);return xx(t)?t.host:t}function u6(e){const t=Ca(e);return Ti(t)?e.ownerDocument?e.ownerDocument.body:e.body:fr(t)&&Su(t)?t:u6(t)}function Tl(e,t,n){var r;t===void 0&&(t=[]),n===void 0&&(n=!0);const a=u6(e),o=a===((r=e.ownerDocument)==null?void 0:r.body),i=nn(a);if(o){const s=zh(i);return t.concat(i,i.visualViewport||[],Su(a)?a:[],s&&n?Tl(s):[])}return t.concat(a,Tl(a,[],n))}function zh(e){return e.parent&&Object.getPrototypeOf(e.parent)?e.frameElement:null}function c6(e){const t=Wn(e);let n=parseFloat(t.width)||0,r=parseFloat(t.height)||0;const a=fr(e),o=a?e.offsetWidth:n,i=a?e.offsetHeight:r,s=sp(n)!==o||sp(r)!==i;return s&&(n=o,r=i),{width:n,height:r,$:s}}function Cy(e){return Un(e)?e:e.contextElement}function ii(e){const t=Cy(e);if(!fr(t))return ur(1);const n=t.getBoundingClientRect(),{width:r,height:a,$:o}=c6(t);let i=(o?sp(n.width):n.width)/r,s=(o?sp(n.height):n.height)/a;return(!i||!Number.isFinite(i))&&(i=1),(!s||!Number.isFinite(s))&&(s=1),{x:i,y:s}}const T_=ur(0);function p6(e){const t=nn(e);return!ky()||!t.visualViewport?T_:{x:t.visualViewport.offsetLeft,y:t.visualViewport.offsetTop}}function j_(e,t,n){return t===void 0&&(t=!1),!n||t&&n!==nn(e)?!1:t}function vo(e,t,n,r){t===void 0&&(t=!1),n===void 0&&(n=!1);const a=e.getBoundingClientRect(),o=Cy(e);let i=ur(1);t&&(r?Un(r)&&(i=ii(r)):i=ii(e));const s=j_(o,n,r)?p6(o):ur(0);let l=(a.left+s.x)/i.x,u=(a.top+s.y)/i.y,p=a.width/i.x,c=a.height/i.y;if(o){const f=nn(o),m=r&&Un(r)?nn(r):r;let h=f,d=zh(h);for(;d&&r&&m!==h;){const g=ii(d),v=d.getBoundingClientRect(),y=Wn(d),x=v.left+(d.clientLeft+parseFloat(y.paddingLeft))*g.x,P=v.top+(d.clientTop+parseFloat(y.paddingTop))*g.y;l*=g.x,u*=g.y,p*=g.x,c*=g.y,l+=x,u+=P,h=nn(d),d=zh(h)}}return up({width:p,height:c,x:l,y:u})}function _y(e,t){const n=Rf(e).scrollLeft;return t?t.left+n:vo(hr(e)).left+n}function f6(e,t,n){n===void 0&&(n=!1);const r=e.getBoundingClientRect(),a=r.left+t.scrollLeft-(n?0:_y(e,r)),o=r.top+t.scrollTop;return{x:a,y:o}}function N_(e){let{elements:t,rect:n,offsetParent:r,strategy:a}=e;const o=a==="fixed",i=hr(r),s=t?Mf(t.floating):!1;if(r===i||s&&o)return n;let l={scrollLeft:0,scrollTop:0},u=ur(1);const p=ur(0),c=fr(r);if((c||!c&&!o)&&((ls(r)!=="body"||Su(i))&&(l=Rf(r)),fr(r))){const m=vo(r);u=ii(r),p.x=m.x+r.clientLeft,p.y=m.y+r.clientTop}const f=i&&!c&&!o?f6(i,l,!0):ur(0);return{width:n.width*u.x,height:n.height*u.y,x:n.x*u.x-l.scrollLeft*u.x+p.x+f.x,y:n.y*u.y-l.scrollTop*u.y+p.y+f.y}}function $_(e){return Array.from(e.getClientRects())}function M_(e){const t=hr(e),n=Rf(e),r=e.ownerDocument.body,a=Qt(t.scrollWidth,t.clientWidth,r.scrollWidth,r.clientWidth),o=Qt(t.scrollHeight,t.clientHeight,r.scrollHeight,r.clientHeight);let i=-n.scrollLeft+_y(e);const s=-n.scrollTop;return Wn(r).direction==="rtl"&&(i+=Qt(t.clientWidth,r.clientWidth)-a),{width:a,height:o,x:i,y:s}}function R_(e,t){const n=nn(e),r=hr(e),a=n.visualViewport;let o=r.clientWidth,i=r.clientHeight,s=0,l=0;if(a){o=a.width,i=a.height;const u=ky();(!u||u&&t==="fixed")&&(s=a.offsetLeft,l=a.offsetTop)}return{width:o,height:i,x:s,y:l}}function I_(e,t){const n=vo(e,!0,t==="fixed"),r=n.top+e.clientTop,a=n.left+e.clientLeft,o=fr(e)?ii(e):ur(1),i=e.clientWidth*o.x,s=e.clientHeight*o.y,l=a*o.x,u=r*o.y;return{width:i,height:s,x:l,y:u}}function wx(e,t,n){let r;if(t==="viewport")r=R_(e,n);else if(t==="document")r=M_(hr(e));else if(Un(t))r=I_(t,n);else{const a=p6(e);r={x:t.x-a.x,y:t.y-a.y,width:t.width,height:t.height}}return up(r)}function d6(e,t){const n=Ca(e);return n===t||!Un(n)||Ti(n)?!1:Wn(n).position==="fixed"||d6(n,t)}function D_(e,t){const n=t.get(e);if(n)return n;let r=Tl(e,[],!1).filter(s=>Un(s)&&ls(s)!=="body"),a=null;const o=Wn(e).position==="fixed";let i=o?Ca(e):e;for(;Un(i)&&!Ti(i);){const s=Wn(i),l=Oy(i);!l&&s.position==="fixed"&&(a=null),(o?!l&&!a:!l&&s.position==="static"&&!!a&&["absolute","fixed"].includes(a.position)||Su(i)&&!l&&d6(e,i))?r=r.filter(p=>p!==i):a=s,i=Ca(i)}return t.set(e,r),r}function L_(e){let{element:t,boundary:n,rootBoundary:r,strategy:a}=e;const i=[...n==="clippingAncestors"?Mf(t)?[]:D_(t,this._c):[].concat(n),r],s=i[0],l=i.reduce((u,p)=>{const c=wx(t,p,a);return u.top=Qt(c.top,u.top),u.right=ka(c.right,u.right),u.bottom=ka(c.bottom,u.bottom),u.left=Qt(c.left,u.left),u},wx(t,s,a));return{width:l.right-l.left,height:l.bottom-l.top,x:l.left,y:l.top}}function F_(e){const{width:t,height:n}=c6(e);return{width:t,height:n}}function B_(e,t,n){const r=fr(t),a=hr(t),o=n==="fixed",i=vo(e,!0,o,t);let s={scrollLeft:0,scrollTop:0};const l=ur(0);function u(){l.x=_y(a)}if(r||!r&&!o)if((ls(t)!=="body"||Su(a))&&(s=Rf(t)),r){const m=vo(t,!0,o,t);l.x=m.x+t.clientLeft,l.y=m.y+t.clientTop}else a&&u();o&&!r&&a&&u();const p=a&&!r&&!o?f6(a,s):ur(0),c=i.left+s.scrollLeft-l.x-p.x,f=i.top+s.scrollTop-l.y-p.y;return{x:c,y:f,width:i.width,height:i.height}}function um(e){return Wn(e).position==="static"}function bx(e,t){if(!fr(e)||Wn(e).position==="fixed")return null;if(t)return t(e);let n=e.offsetParent;return hr(e)===n&&(n=n.ownerDocument.body),n}function m6(e,t){const n=nn(e);if(Mf(e))return n;if(!fr(e)){let a=Ca(e);for(;a&&!Ti(a);){if(Un(a)&&!um(a))return a;a=Ca(a)}return n}let r=bx(e,t);for(;r&&A_(r)&&um(r);)r=bx(r,t);return r&&Ti(r)&&um(r)&&!Oy(r)?n:r||E_(e)||n}const z_=async function(e){const t=this.getOffsetParent||m6,n=this.getDimensions,r=await n(e.floating);return{reference:B_(e.reference,await t(e.floating),e.strategy),floating:{x:0,y:0,width:r.width,height:r.height}}};function H_(e){return Wn(e).direction==="rtl"}const G_={convertOffsetParentRelativeRectToViewportRelativeRect:N_,getDocumentElement:hr,getClippingRect:L_,getOffsetParent:m6,getElementRects:z_,getClientRects:$_,getDimensions:F_,getScale:ii,isElement:Un,isRTL:H_};function h6(e,t){return e.x===t.x&&e.y===t.y&&e.width===t.width&&e.height===t.height}function U_(e,t){let n=null,r;const a=hr(e);function o(){var s;clearTimeout(r),(s=n)==null||s.disconnect(),n=null}function i(s,l){s===void 0&&(s=!1),l===void 0&&(l=1),o();const u=e.getBoundingClientRect(),{left:p,top:c,width:f,height:m}=u;if(s||t(),!f||!m)return;const h=Zu(c),d=Zu(a.clientWidth-(p+f)),g=Zu(a.clientHeight-(c+m)),v=Zu(p),x={rootMargin:-h+"px "+-d+"px "+-g+"px "+-v+"px",threshold:Qt(0,ka(1,l))||1};let P=!0;function w(S){const O=S[0].intersectionRatio;if(O!==l){if(!P)return i();O?i(!1,O):r=setTimeout(()=>{i(!1,1e-7)},1e3)}O===1&&!h6(u,e.getBoundingClientRect())&&i(),P=!1}try{n=new IntersectionObserver(w,{...x,root:a.ownerDocument})}catch{n=new IntersectionObserver(w,x)}n.observe(e)}return i(!0),o}function W_(e,t,n,r){r===void 0&&(r={});const{ancestorScroll:a=!0,ancestorResize:o=!0,elementResize:i=typeof ResizeObserver=="function",layoutShift:s=typeof IntersectionObserver=="function",animationFrame:l=!1}=r,u=Cy(e),p=a||o?[...u?Tl(u):[],...Tl(t)]:[];p.forEach(v=>{a&&v.addEventListener("scroll",n,{passive:!0}),o&&v.addEventListener("resize",n)});const c=u&&s?U_(u,n):null;let f=-1,m=null;i&&(m=new ResizeObserver(v=>{let[y]=v;y&&y.target===u&&m&&(m.unobserve(t),cancelAnimationFrame(f),f=requestAnimationFrame(()=>{var x;(x=m)==null||x.observe(t)})),n()}),u&&!l&&m.observe(u),m.observe(t));let h,d=l?vo(e):null;l&&g();function g(){const v=vo(e);d&&!h6(d,v)&&n(),d=v,h=requestAnimationFrame(g)}return n(),()=>{var v;p.forEach(y=>{a&&y.removeEventListener("scroll",n),o&&y.removeEventListener("resize",n)}),c==null||c(),(v=m)==null||v.disconnect(),m=null,l&&cancelAnimationFrame(h)}}const q_=O_,V_=k_,K_=b_,X_=__,Y_=P_,Px=w_,Q_=C_,Z_=(e,t,n)=>{const r=new Map,a={platform:G_,...n},o={...a.platform,_c:r};return x_(e,t,{...a,platform:o})};var J_=typeof document<"u",eA=function(){},jc=J_?k.useLayoutEffect:eA;function cp(e,t){if(e===t)return!0;if(typeof e!=typeof t)return!1;if(typeof e=="function"&&e.toString()===t.toString())return!0;let n,r,a;if(e&&t&&typeof e=="object"){if(Array.isArray(e)){if(n=e.length,n!==t.length)return!1;for(r=n;r--!==0;)if(!cp(e[r],t[r]))return!1;return!0}if(a=Object.keys(e),n=a.length,n!==Object.keys(t).length)return!1;for(r=n;r--!==0;)if(!{}.hasOwnProperty.call(t,a[r]))return!1;for(r=n;r--!==0;){const o=a[r];if(!(o==="_owner"&&e.$$typeof)&&!cp(e[o],t[o]))return!1}return!0}return e!==e&&t!==t}function v6(e){return typeof window>"u"?1:(e.ownerDocument.defaultView||window).devicePixelRatio||1}function Sx(e,t){const n=v6(e);return Math.round(t*n)/n}function cm(e){const t=k.useRef(e);return jc(()=>{t.current=e}),t}function tA(e){e===void 0&&(e={});const{placement:t="bottom",strategy:n="absolute",middleware:r=[],platform:a,elements:{reference:o,floating:i}={},transform:s=!0,whileElementsMounted:l,open:u}=e,[p,c]=k.useState({x:0,y:0,strategy:n,placement:t,middlewareData:{},isPositioned:!1}),[f,m]=k.useState(r);cp(f,r)||m(r);const[h,d]=k.useState(null),[g,v]=k.useState(null),y=k.useCallback(N=>{N!==S.current&&(S.current=N,d(N))},[]),x=k.useCallback(N=>{N!==O.current&&(O.current=N,v(N))},[]),P=o||h,w=i||g,S=k.useRef(null),O=k.useRef(null),C=k.useRef(p),_=l!=null,T=cm(l),A=cm(a),j=cm(u),$=k.useCallback(()=>{if(!S.current||!O.current)return;const N={placement:t,strategy:n,middleware:f};A.current&&(N.platform=A.current),Z_(S.current,O.current,N).then(D=>{const H={...D,isPositioned:j.current!==!1};M.current&&!cp(C.current,H)&&(C.current=H,os.flushSync(()=>{c(H)}))})},[f,t,n,A,j]);jc(()=>{u===!1&&C.current.isPositioned&&(C.current.isPositioned=!1,c(N=>({...N,isPositioned:!1})))},[u]);const M=k.useRef(!1);jc(()=>(M.current=!0,()=>{M.current=!1}),[]),jc(()=>{if(P&&(S.current=P),w&&(O.current=w),P&&w){if(T.current)return T.current(P,w,$);$()}},[P,w,$,T,_]);const I=k.useMemo(()=>({reference:S,floating:O,setReference:y,setFloating:x}),[y,x]),R=k.useMemo(()=>({reference:P,floating:w}),[P,w]),L=k.useMemo(()=>{const N={position:n,left:0,top:0};if(!R.floating)return N;const D=Sx(R.floating,p.x),H=Sx(R.floating,p.y);return s?{...N,transform:"translate("+D+"px, "+H+"px)",...v6(R.floating)>=1.5&&{willChange:"transform"}}:{position:n,left:D,top:H}},[n,s,R.floating,p.x,p.y]);return k.useMemo(()=>({...p,update:$,refs:I,elements:R,floatingStyles:L}),[p,$,I,R,L])}const nA=e=>{function t(n){return{}.hasOwnProperty.call(n,"current")}return{name:"arrow",options:e,fn(n){const{element:r,padding:a}=typeof e=="function"?e(n):e;return r&&t(r)?r.current!=null?Px({element:r.current,padding:a}).fn(n):{}:r?Px({element:r,padding:a}).fn(n):{}}}},rA=(e,t)=>({...q_(e),options:[e,t]}),aA=(e,t)=>({...V_(e),options:[e,t]}),oA=(e,t)=>({...Q_(e),options:[e,t]}),iA=(e,t)=>({...K_(e),options:[e,t]}),sA=(e,t)=>({...X_(e),options:[e,t]}),lA=(e,t)=>({...Y_(e),options:[e,t]}),uA=(e,t)=>({...nA(e),options:[e,t]});var cA="Arrow",y6=k.forwardRef((e,t)=>{const{children:n,width:r=10,height:a=5,...o}=e;return b.jsx(Ae.svg,{...o,ref:t,width:r,height:a,viewBox:"0 0 30 10",preserveAspectRatio:"none",children:e.asChild?n:b.jsx("polygon",{points:"0,0 30,0 15,10"})})});y6.displayName=cA;var pA=y6;function fA(e){const[t,n]=k.useState(void 0);return Et(()=>{if(e){n({width:e.offsetWidth,height:e.offsetHeight});const r=new ResizeObserver(a=>{if(!Array.isArray(a)||!a.length)return;const o=a[0];let i,s;if("borderBoxSize"in o){const l=o.borderBoxSize,u=Array.isArray(l)?l[0]:l;i=u.inlineSize,s=u.blockSize}else i=e.offsetWidth,s=e.offsetHeight;n({width:i,height:s})});return r.observe(e,{box:"border-box"}),()=>r.unobserve(e)}else n(void 0)},[e]),t}var Ay="Popper",[g6,If]=is(Ay),[dA,x6]=g6(Ay),w6=e=>{const{__scopePopper:t,children:n}=e,[r,a]=k.useState(null);return b.jsx(dA,{scope:t,anchor:r,onAnchorChange:a,children:n})};w6.displayName=Ay;var b6="PopperAnchor",P6=k.forwardRef((e,t)=>{const{__scopePopper:n,virtualRef:r,...a}=e,o=x6(b6,n),i=k.useRef(null),s=Xe(t,i);return k.useEffect(()=>{o.onAnchorChange((r==null?void 0:r.current)||i.current)}),r?null:b.jsx(Ae.div,{...a,ref:s})});P6.displayName=b6;var Ey="PopperContent",[mA,hA]=g6(Ey),S6=k.forwardRef((e,t)=>{var J,se,q,K,X,F;const{__scopePopper:n,side:r="bottom",sideOffset:a=0,align:o="center",alignOffset:i=0,arrowPadding:s=0,avoidCollisions:l=!0,collisionBoundary:u=[],collisionPadding:p=0,sticky:c="partial",hideWhenDetached:f=!1,updatePositionStrategy:m="optimized",onPlaced:h,...d}=e,g=x6(Ey,n),[v,y]=k.useState(null),x=Xe(t,pe=>y(pe)),[P,w]=k.useState(null),S=fA(P),O=(S==null?void 0:S.width)??0,C=(S==null?void 0:S.height)??0,_=r+(o!=="center"?"-"+o:""),T=typeof p=="number"?p:{top:0,right:0,bottom:0,left:0,...p},A=Array.isArray(u)?u:[u],j=A.length>0,$={padding:T,boundary:A.filter(yA),altBoundary:j},{refs:M,floatingStyles:I,placement:R,isPositioned:L,middlewareData:N}=tA({strategy:"fixed",placement:_,whileElementsMounted:(...pe)=>W_(...pe,{animationFrame:m==="always"}),elements:{reference:g.anchor},middleware:[rA({mainAxis:a+C,alignmentAxis:i}),l&&aA({mainAxis:!0,crossAxis:!1,limiter:c==="partial"?oA():void 0,...$}),l&&iA({...$}),sA({...$,apply:({elements:pe,rects:te,availableWidth:$e,availableHeight:Me})=>{const{width:Qe,height:Vn}=te.reference,Sn=pe.floating.style;Sn.setProperty("--radix-popper-available-width",`${$e}px`),Sn.setProperty("--radix-popper-available-height",`${Me}px`),Sn.setProperty("--radix-popper-anchor-width",`${Qe}px`),Sn.setProperty("--radix-popper-anchor-height",`${Vn}px`)}}),P&&uA({element:P,padding:s}),gA({arrowWidth:O,arrowHeight:C}),f&&lA({strategy:"referenceHidden",...$})]}),[D,H]=C6(R),W=Oa(h);Et(()=>{L&&(W==null||W())},[L,W]);const G=(J=N.arrow)==null?void 0:J.x,Z=(se=N.arrow)==null?void 0:se.y,re=((q=N.arrow)==null?void 0:q.centerOffset)!==0,[ve,be]=k.useState();return Et(()=>{v&&be(window.getComputedStyle(v).zIndex)},[v]),b.jsx("div",{ref:M.setFloating,"data-radix-popper-content-wrapper":"",style:{...I,transform:L?I.transform:"translate(0, -200%)",minWidth:"max-content",zIndex:ve,"--radix-popper-transform-origin":[(K=N.transformOrigin)==null?void 0:K.x,(X=N.transformOrigin)==null?void 0:X.y].join(" "),...((F=N.hide)==null?void 0:F.referenceHidden)&&{visibility:"hidden",pointerEvents:"none"}},dir:e.dir,children:b.jsx(mA,{scope:n,placedSide:D,onArrowChange:w,arrowX:G,arrowY:Z,shouldHideArrow:re,children:b.jsx(Ae.div,{"data-side":D,"data-align":H,...d,ref:x,style:{...d.style,animation:L?void 0:"none"}})})})});S6.displayName=Ey;var O6="PopperArrow",vA={top:"bottom",right:"left",bottom:"top",left:"right"},k6=k.forwardRef(function(t,n){const{__scopePopper:r,...a}=t,o=hA(O6,r),i=vA[o.placedSide];return b.jsx("span",{ref:o.onArrowChange,style:{position:"absolute",left:o.arrowX,top:o.arrowY,[i]:0,transformOrigin:{top:"",right:"0 0",bottom:"center 0",left:"100% 0"}[o.placedSide],transform:{top:"translateY(100%)",right:"translateY(50%) rotate(90deg) translateX(-50%)",bottom:"rotate(180deg)",left:"translateY(50%) rotate(-90deg) translateX(50%)"}[o.placedSide],visibility:o.shouldHideArrow?"hidden":void 0},children:b.jsx(pA,{...a,ref:n,style:{...a.style,display:"block"}})})});k6.displayName=O6;function yA(e){return e!==null}var gA=e=>({name:"transformOrigin",options:e,fn(t){var g,v,y;const{placement:n,rects:r,middlewareData:a}=t,i=((g=a.arrow)==null?void 0:g.centerOffset)!==0,s=i?0:e.arrowWidth,l=i?0:e.arrowHeight,[u,p]=C6(n),c={start:"0%",center:"50%",end:"100%"}[p],f=(((v=a.arrow)==null?void 0:v.x)??0)+s/2,m=(((y=a.arrow)==null?void 0:y.y)??0)+l/2;let h="",d="";return u==="bottom"?(h=i?c:`${f}px`,d=`${-l}px`):u==="top"?(h=i?c:`${f}px`,d=`${r.floating.height+l}px`):u==="right"?(h=`${-l}px`,d=i?c:`${m}px`):u==="left"&&(h=`${r.floating.width+l}px`,d=i?c:`${m}px`),{data:{x:h,y:d}}}});function C6(e){const[t,n="center"]=e.split("-");return[t,n]}var xA=w6,_6=P6,A6=S6,E6=k6,wA="Portal",T6=k.forwardRef((e,t)=>{var s;const{container:n,...r}=e,[a,o]=k.useState(!1);Et(()=>o(!0),[]);const i=n||a&&((s=globalThis==null?void 0:globalThis.document)==null?void 0:s.body);return i?IC.createPortal(b.jsx(Ae.div,{...r,ref:t}),i):null});T6.displayName=wA;var bA=Q2[" useInsertionEffect ".trim().toString()]||Et;function pp({prop:e,defaultProp:t,onChange:n=()=>{},caller:r}){const[a,o,i]=PA({defaultProp:t,onChange:n}),s=e!==void 0,l=s?e:a;{const p=k.useRef(e!==void 0);k.useEffect(()=>{const c=p.current;c!==s&&console.warn(`${r} is changing from ${c?"controlled":"uncontrolled"} to ${s?"controlled":"uncontrolled"}. Components should not switch from controlled to uncontrolled (or vice versa). Decide between using a controlled or uncontrolled value for the lifetime of the component.`),p.current=s},[s,r])}const u=k.useCallback(p=>{var c;if(s){const f=SA(p)?p(e):p;f!==e&&((c=i.current)==null||c.call(i,f))}else o(p)},[s,e,o,i]);return[l,u]}function PA({defaultProp:e,onChange:t}){const[n,r]=k.useState(e),a=k.useRef(n),o=k.useRef(t);return bA(()=>{o.current=t},[t]),k.useEffect(()=>{var i;a.current!==n&&((i=o.current)==null||i.call(o,n),a.current=n)},[n,a]),[n,r,o]}function SA(e){return typeof e=="function"}function OA(e){const t=k.useRef({value:e,previous:e});return k.useMemo(()=>(t.current.value!==e&&(t.current.previous=t.current.value,t.current.value=e),t.current.previous),[e])}var j6=Object.freeze({position:"absolute",border:0,width:1,height:1,padding:0,margin:-1,overflow:"hidden",clip:"rect(0, 0, 0, 0)",whiteSpace:"nowrap",wordWrap:"normal"}),kA="VisuallyHidden",N6=k.forwardRef((e,t)=>b.jsx(Ae.span,{...e,ref:t,style:{...j6,...e.style}}));N6.displayName=kA;var CA=N6,_A=function(e){if(typeof document>"u")return null;var t=Array.isArray(e)?e[0]:e;return t.ownerDocument.body},No=new WeakMap,Ju=new WeakMap,ec={},pm=0,$6=function(e){return e&&(e.host||$6(e.parentNode))},AA=function(e,t){return t.map(function(n){if(e.contains(n))return n;var r=$6(n);return r&&e.contains(r)?r:(console.error("aria-hidden",n,"in not contained inside",e,". Doing nothing"),null)}).filter(function(n){return!!n})},EA=function(e,t,n,r){var a=AA(t,Array.isArray(e)?e:[e]);ec[n]||(ec[n]=new WeakMap);var o=ec[n],i=[],s=new Set,l=new Set(a),u=function(c){!c||s.has(c)||(s.add(c),u(c.parentNode))};a.forEach(u);var p=function(c){!c||l.has(c)||Array.prototype.forEach.call(c.children,function(f){if(s.has(f))p(f);else try{var m=f.getAttribute(r),h=m!==null&&m!=="false",d=(No.get(f)||0)+1,g=(o.get(f)||0)+1;No.set(f,d),o.set(f,g),i.push(f),d===1&&h&&Ju.set(f,!0),g===1&&f.setAttribute(n,"true"),h||f.setAttribute(r,"true")}catch(v){console.error("aria-hidden: cannot operate on ",f,v)}})};return p(t),s.clear(),pm++,function(){i.forEach(function(c){var f=No.get(c)-1,m=o.get(c)-1;No.set(c,f),o.set(c,m),f||(Ju.has(c)||c.removeAttribute(r),Ju.delete(c)),m||c.removeAttribute(n)}),pm--,pm||(No=new WeakMap,No=new WeakMap,Ju=new WeakMap,ec={})}},TA=function(e,t,n){n===void 0&&(n="data-aria-hidden");var r=Array.from(Array.isArray(e)?e:[e]),a=_A(e);return a?(r.push.apply(r,Array.from(a.querySelectorAll("[aria-live], script"))),EA(r,a,n,"aria-hidden")):function(){return null}},tr=function(){return tr=Object.assign||function(t){for(var n,r=1,a=arguments.length;r<a;r++){n=arguments[r];for(var o in n)Object.prototype.hasOwnProperty.call(n,o)&&(t[o]=n[o])}return t},tr.apply(this,arguments)};function M6(e,t){var n={};for(var r in e)Object.prototype.hasOwnProperty.call(e,r)&&t.indexOf(r)<0&&(n[r]=e[r]);if(e!=null&&typeof Object.getOwnPropertySymbols=="function")for(var a=0,r=Object.getOwnPropertySymbols(e);a<r.length;a++)t.indexOf(r[a])<0&&Object.prototype.propertyIsEnumerable.call(e,r[a])&&(n[r[a]]=e[r[a]]);return n}function jA(e,t,n){if(n||arguments.length===2)for(var r=0,a=t.length,o;r<a;r++)(o||!(r in t))&&(o||(o=Array.prototype.slice.call(t,0,r)),o[r]=t[r]);return e.concat(o||Array.prototype.slice.call(t))}var Nc="right-scroll-bar-position",$c="width-before-scroll-bar",NA="with-scroll-bars-hidden",$A="--removed-body-scroll-bar-size";function fm(e,t){return typeof e=="function"?e(t):e&&(e.current=t),e}function MA(e,t){var n=k.useState(function(){return{value:e,callback:t,facade:{get current(){return n.value},set current(r){var a=n.value;a!==r&&(n.value=r,n.callback(r,a))}}}})[0];return n.callback=t,n.facade}var RA=typeof window<"u"?k.useLayoutEffect:k.useEffect,Ox=new WeakMap;function IA(e,t){var n=MA(null,function(r){return e.forEach(function(a){return fm(a,r)})});return RA(function(){var r=Ox.get(n);if(r){var a=new Set(r),o=new Set(e),i=n.current;a.forEach(function(s){o.has(s)||fm(s,null)}),o.forEach(function(s){a.has(s)||fm(s,i)})}Ox.set(n,e)},[e]),n}function DA(e){return e}function LA(e,t){t===void 0&&(t=DA);var n=[],r=!1,a={read:function(){if(r)throw new Error("Sidecar: could not `read` from an `assigned` medium. `read` could be used only with `useMedium`.");return n.length?n[n.length-1]:e},useMedium:function(o){var i=t(o,r);return n.push(i),function(){n=n.filter(function(s){return s!==i})}},assignSyncMedium:function(o){for(r=!0;n.length;){var i=n;n=[],i.forEach(o)}n={push:function(s){return o(s)},filter:function(){return n}}},assignMedium:function(o){r=!0;var i=[];if(n.length){var s=n;n=[],s.forEach(o),i=n}var l=function(){var p=i;i=[],p.forEach(o)},u=function(){return Promise.resolve().then(l)};u(),n={push:function(p){i.push(p),u()},filter:function(p){return i=i.filter(p),n}}}};return a}function FA(e){e===void 0&&(e={});var t=LA(null);return t.options=tr({async:!0,ssr:!1},e),t}var R6=function(e){var t=e.sideCar,n=M6(e,["sideCar"]);if(!t)throw new Error("Sidecar: please provide `sideCar` property to import the right car");var r=t.read();if(!r)throw new Error("Sidecar medium not found");return k.createElement(r,tr({},n))};R6.isSideCarExport=!0;function BA(e,t){return e.useMedium(t),R6}var I6=FA(),dm=function(){},Df=k.forwardRef(function(e,t){var n=k.useRef(null),r=k.useState({onScrollCapture:dm,onWheelCapture:dm,onTouchMoveCapture:dm}),a=r[0],o=r[1],i=e.forwardProps,s=e.children,l=e.className,u=e.removeScrollBar,p=e.enabled,c=e.shards,f=e.sideCar,m=e.noRelative,h=e.noIsolation,d=e.inert,g=e.allowPinchZoom,v=e.as,y=v===void 0?"div":v,x=e.gapMode,P=M6(e,["forwardProps","children","className","removeScrollBar","enabled","shards","sideCar","noRelative","noIsolation","inert","allowPinchZoom","as","gapMode"]),w=f,S=IA([n,t]),O=tr(tr({},P),a);return k.createElement(k.Fragment,null,p&&k.createElement(w,{sideCar:I6,removeScrollBar:u,shards:c,noRelative:m,noIsolation:h,inert:d,setCallbacks:o,allowPinchZoom:!!g,lockRef:n,gapMode:x}),i?k.cloneElement(k.Children.only(s),tr(tr({},O),{ref:S})):k.createElement(y,tr({},O,{className:l,ref:S}),s))});Df.defaultProps={enabled:!0,removeScrollBar:!0,inert:!1};Df.classNames={fullWidth:$c,zeroRight:Nc};var zA=function(){if(typeof __webpack_nonce__<"u")return __webpack_nonce__};function HA(){if(!document)return null;var e=document.createElement("style");e.type="text/css";var t=zA();return t&&e.setAttribute("nonce",t),e}function GA(e,t){e.styleSheet?e.styleSheet.cssText=t:e.appendChild(document.createTextNode(t))}function UA(e){var t=document.head||document.getElementsByTagName("head")[0];t.appendChild(e)}var WA=function(){var e=0,t=null;return{add:function(n){e==0&&(t=HA())&&(GA(t,n),UA(t)),e++},remove:function(){e--,!e&&t&&(t.parentNode&&t.parentNode.removeChild(t),t=null)}}},qA=function(){var e=WA();return function(t,n){k.useEffect(function(){return e.add(t),function(){e.remove()}},[t&&n])}},D6=function(){var e=qA(),t=function(n){var r=n.styles,a=n.dynamic;return e(r,a),null};return t},VA={left:0,top:0,right:0,gap:0},mm=function(e){return parseInt(e||"",10)||0},KA=function(e){var t=window.getComputedStyle(document.body),n=t[e==="padding"?"paddingLeft":"marginLeft"],r=t[e==="padding"?"paddingTop":"marginTop"],a=t[e==="padding"?"paddingRight":"marginRight"];return[mm(n),mm(r),mm(a)]},XA=function(e){if(e===void 0&&(e="margin"),typeof window>"u")return VA;var t=KA(e),n=document.documentElement.clientWidth,r=window.innerWidth;return{left:t[0],top:t[1],right:t[2],gap:Math.max(0,r-n+t[2]-t[0])}},YA=D6(),si="data-scroll-locked",QA=function(e,t,n,r){var a=e.left,o=e.top,i=e.right,s=e.gap;return n===void 0&&(n="margin"),`
  .`.concat(NA,` {
   overflow: hidden `).concat(r,`;
   padding-right: `).concat(s,"px ").concat(r,`;
  }
  body[`).concat(si,`] {
    overflow: hidden `).concat(r,`;
    overscroll-behavior: contain;
    `).concat([t&&"position: relative ".concat(r,";"),n==="margin"&&`
    padding-left: `.concat(a,`px;
    padding-top: `).concat(o,`px;
    padding-right: `).concat(i,`px;
    margin-left:0;
    margin-top:0;
    margin-right: `).concat(s,"px ").concat(r,`;
    `),n==="padding"&&"padding-right: ".concat(s,"px ").concat(r,";")].filter(Boolean).join(""),`
  }
  
  .`).concat(Nc,` {
    right: `).concat(s,"px ").concat(r,`;
  }
  
  .`).concat($c,` {
    margin-right: `).concat(s,"px ").concat(r,`;
  }
  
  .`).concat(Nc," .").concat(Nc,` {
    right: 0 `).concat(r,`;
  }
  
  .`).concat($c," .").concat($c,` {
    margin-right: 0 `).concat(r,`;
  }
  
  body[`).concat(si,`] {
    `).concat($A,": ").concat(s,`px;
  }
`)},kx=function(){var e=parseInt(document.body.getAttribute(si)||"0",10);return isFinite(e)?e:0},ZA=function(){k.useEffect(function(){return document.body.setAttribute(si,(kx()+1).toString()),function(){var e=kx()-1;e<=0?document.body.removeAttribute(si):document.body.setAttribute(si,e.toString())}},[])},JA=function(e){var t=e.noRelative,n=e.noImportant,r=e.gapMode,a=r===void 0?"margin":r;ZA();var o=k.useMemo(function(){return XA(a)},[a]);return k.createElement(YA,{styles:QA(o,!t,a,n?"":"!important")})},Hh=!1;if(typeof window<"u")try{var tc=Object.defineProperty({},"passive",{get:function(){return Hh=!0,!0}});window.addEventListener("test",tc,tc),window.removeEventListener("test",tc,tc)}catch{Hh=!1}var $o=Hh?{passive:!1}:!1,eE=function(e){return e.tagName==="TEXTAREA"},L6=function(e,t){if(!(e instanceof Element))return!1;var n=window.getComputedStyle(e);return n[t]!=="hidden"&&!(n.overflowY===n.overflowX&&!eE(e)&&n[t]==="visible")},tE=function(e){return L6(e,"overflowY")},nE=function(e){return L6(e,"overflowX")},Cx=function(e,t){var n=t.ownerDocument,r=t;do{typeof ShadowRoot<"u"&&r instanceof ShadowRoot&&(r=r.host);var a=F6(e,r);if(a){var o=B6(e,r),i=o[1],s=o[2];if(i>s)return!0}r=r.parentNode}while(r&&r!==n.body);return!1},rE=function(e){var t=e.scrollTop,n=e.scrollHeight,r=e.clientHeight;return[t,n,r]},aE=function(e){var t=e.scrollLeft,n=e.scrollWidth,r=e.clientWidth;return[t,n,r]},F6=function(e,t){return e==="v"?tE(t):nE(t)},B6=function(e,t){return e==="v"?rE(t):aE(t)},oE=function(e,t){return e==="h"&&t==="rtl"?-1:1},iE=function(e,t,n,r,a){var o=oE(e,window.getComputedStyle(t).direction),i=o*r,s=n.target,l=t.contains(s),u=!1,p=i>0,c=0,f=0;do{if(!s)break;var m=B6(e,s),h=m[0],d=m[1],g=m[2],v=d-g-o*h;(h||v)&&F6(e,s)&&(c+=v,f+=h);var y=s.parentNode;s=y&&y.nodeType===Node.DOCUMENT_FRAGMENT_NODE?y.host:y}while(!l&&s!==document.body||l&&(t.contains(s)||t===s));return(p&&Math.abs(c)<1||!p&&Math.abs(f)<1)&&(u=!0),u},nc=function(e){return"changedTouches"in e?[e.changedTouches[0].clientX,e.changedTouches[0].clientY]:[0,0]},_x=function(e){return[e.deltaX,e.deltaY]},Ax=function(e){return e&&"current"in e?e.current:e},sE=function(e,t){return e[0]===t[0]&&e[1]===t[1]},lE=function(e){return`
  .block-interactivity-`.concat(e,` {pointer-events: none;}
  .allow-interactivity-`).concat(e,` {pointer-events: all;}
`)},uE=0,Mo=[];function cE(e){var t=k.useRef([]),n=k.useRef([0,0]),r=k.useRef(),a=k.useState(uE++)[0],o=k.useState(D6)[0],i=k.useRef(e);k.useEffect(function(){i.current=e},[e]),k.useEffect(function(){if(e.inert){document.body.classList.add("block-interactivity-".concat(a));var d=jA([e.lockRef.current],(e.shards||[]).map(Ax),!0).filter(Boolean);return d.forEach(function(g){return g.classList.add("allow-interactivity-".concat(a))}),function(){document.body.classList.remove("block-interactivity-".concat(a)),d.forEach(function(g){return g.classList.remove("allow-interactivity-".concat(a))})}}},[e.inert,e.lockRef.current,e.shards]);var s=k.useCallback(function(d,g){if("touches"in d&&d.touches.length===2||d.type==="wheel"&&d.ctrlKey)return!i.current.allowPinchZoom;var v=nc(d),y=n.current,x="deltaX"in d?d.deltaX:y[0]-v[0],P="deltaY"in d?d.deltaY:y[1]-v[1],w,S=d.target,O=Math.abs(x)>Math.abs(P)?"h":"v";if("touches"in d&&O==="h"&&S.type==="range")return!1;var C=Cx(O,S);if(!C)return!0;if(C?w=O:(w=O==="v"?"h":"v",C=Cx(O,S)),!C)return!1;if(!r.current&&"changedTouches"in d&&(x||P)&&(r.current=w),!w)return!0;var _=r.current||w;return iE(_,g,d,_==="h"?x:P)},[]),l=k.useCallback(function(d){var g=d;if(!(!Mo.length||Mo[Mo.length-1]!==o)){var v="deltaY"in g?_x(g):nc(g),y=t.current.filter(function(w){return w.name===g.type&&(w.target===g.target||g.target===w.shadowParent)&&sE(w.delta,v)})[0];if(y&&y.should){g.cancelable&&g.preventDefault();return}if(!y){var x=(i.current.shards||[]).map(Ax).filter(Boolean).filter(function(w){return w.contains(g.target)}),P=x.length>0?s(g,x[0]):!i.current.noIsolation;P&&g.cancelable&&g.preventDefault()}}},[]),u=k.useCallback(function(d,g,v,y){var x={name:d,delta:g,target:v,should:y,shadowParent:pE(v)};t.current.push(x),setTimeout(function(){t.current=t.current.filter(function(P){return P!==x})},1)},[]),p=k.useCallback(function(d){n.current=nc(d),r.current=void 0},[]),c=k.useCallback(function(d){u(d.type,_x(d),d.target,s(d,e.lockRef.current))},[]),f=k.useCallback(function(d){u(d.type,nc(d),d.target,s(d,e.lockRef.current))},[]);k.useEffect(function(){return Mo.push(o),e.setCallbacks({onScrollCapture:c,onWheelCapture:c,onTouchMoveCapture:f}),document.addEventListener("wheel",l,$o),document.addEventListener("touchmove",l,$o),document.addEventListener("touchstart",p,$o),function(){Mo=Mo.filter(function(d){return d!==o}),document.removeEventListener("wheel",l,$o),document.removeEventListener("touchmove",l,$o),document.removeEventListener("touchstart",p,$o)}},[]);var m=e.removeScrollBar,h=e.inert;return k.createElement(k.Fragment,null,h?k.createElement(o,{styles:lE(a)}):null,m?k.createElement(JA,{noRelative:e.noRelative,gapMode:e.gapMode}):null)}function pE(e){for(var t=null;e!==null;)e instanceof ShadowRoot&&(t=e.host,e=e.host),e=e.parentNode;return t}const fE=BA(I6,cE);var z6=k.forwardRef(function(e,t){return k.createElement(Df,tr({},e,{ref:t,sideCar:fE}))});z6.classNames=Df.classNames;var dE=[" ","Enter","ArrowUp","ArrowDown"],mE=[" ","Enter"],yo="Select",[Lf,Ff,hE]=n6(yo),[us,Cie]=is(yo,[hE,If]),Bf=If(),[vE,Ta]=us(yo),[yE,gE]=us(yo),H6=e=>{const{__scopeSelect:t,children:n,open:r,defaultOpen:a,onOpenChange:o,value:i,defaultValue:s,onValueChange:l,dir:u,name:p,autoComplete:c,disabled:f,required:m,form:h}=e,d=Bf(t),[g,v]=k.useState(null),[y,x]=k.useState(null),[P,w]=k.useState(!1),S=xy(u),[O,C]=pp({prop:r,defaultProp:a??!1,onChange:o,caller:yo}),[_,T]=pp({prop:i,defaultProp:s,onChange:l,caller:yo}),A=k.useRef(null),j=g?h||!!g.closest("form"):!0,[$,M]=k.useState(new Set),I=Array.from($).map(R=>R.props.value).join(";");return b.jsx(xA,{...d,children:b.jsxs(vE,{required:m,scope:t,trigger:g,onTriggerChange:v,valueNode:y,onValueNodeChange:x,valueNodeHasChildren:P,onValueNodeHasChildrenChange:w,contentId:Pu(),value:_,onValueChange:T,open:O,onOpenChange:C,dir:S,triggerPointerDownPosRef:A,disabled:f,children:[b.jsx(Lf.Provider,{scope:t,children:b.jsx(yE,{scope:e.__scopeSelect,onNativeOptionAdd:k.useCallback(R=>{M(L=>new Set(L).add(R))},[]),onNativeOptionRemove:k.useCallback(R=>{M(L=>{const N=new Set(L);return N.delete(R),N})},[]),children:n})}),j?b.jsxs(f4,{"aria-hidden":!0,required:m,tabIndex:-1,name:p,autoComplete:c,value:_,onChange:R=>T(R.target.value),disabled:f,form:h,children:[_===void 0?b.jsx("option",{value:""}):null,Array.from($)]},I):null]})})};H6.displayName=yo;var G6="SelectTrigger",U6=k.forwardRef((e,t)=>{const{__scopeSelect:n,disabled:r=!1,...a}=e,o=Bf(n),i=Ta(G6,n),s=i.disabled||r,l=Xe(t,i.onTriggerChange),u=Ff(n),p=k.useRef("touch"),[c,f,m]=m4(d=>{const g=u().filter(x=>!x.disabled),v=g.find(x=>x.value===i.value),y=h4(g,d,v);y!==void 0&&i.onValueChange(y.value)}),h=d=>{s||(i.onOpenChange(!0),m()),d&&(i.triggerPointerDownPosRef.current={x:Math.round(d.pageX),y:Math.round(d.pageY)})};return b.jsx(_6,{asChild:!0,...o,children:b.jsx(Ae.button,{type:"button",role:"combobox","aria-controls":i.contentId,"aria-expanded":i.open,"aria-required":i.required,"aria-autocomplete":"none",dir:i.dir,"data-state":i.open?"open":"closed",disabled:s,"data-disabled":s?"":void 0,"data-placeholder":d4(i.value)?"":void 0,...a,ref:l,onClick:fe(a.onClick,d=>{d.currentTarget.focus(),p.current!=="mouse"&&h(d)}),onPointerDown:fe(a.onPointerDown,d=>{p.current=d.pointerType;const g=d.target;g.hasPointerCapture(d.pointerId)&&g.releasePointerCapture(d.pointerId),d.button===0&&d.ctrlKey===!1&&d.pointerType==="mouse"&&(h(d),d.preventDefault())}),onKeyDown:fe(a.onKeyDown,d=>{const g=c.current!=="";!(d.ctrlKey||d.altKey||d.metaKey)&&d.key.length===1&&f(d.key),!(g&&d.key===" ")&&dE.includes(d.key)&&(h(),d.preventDefault())})})})});U6.displayName=G6;var W6="SelectValue",q6=k.forwardRef((e,t)=>{const{__scopeSelect:n,className:r,style:a,children:o,placeholder:i="",...s}=e,l=Ta(W6,n),{onValueNodeHasChildrenChange:u}=l,p=o!==void 0,c=Xe(t,l.onValueNodeChange);return Et(()=>{u(p)},[u,p]),b.jsx(Ae.span,{...s,ref:c,style:{pointerEvents:"none"},children:d4(l.value)?b.jsx(b.Fragment,{children:i}):o})});q6.displayName=W6;var xE="SelectIcon",V6=k.forwardRef((e,t)=>{const{__scopeSelect:n,children:r,...a}=e;return b.jsx(Ae.span,{"aria-hidden":!0,...a,ref:t,children:r||"▼"})});V6.displayName=xE;var wE="SelectPortal",K6=e=>b.jsx(T6,{asChild:!0,...e});K6.displayName=wE;var go="SelectContent",X6=k.forwardRef((e,t)=>{const n=Ta(go,e.__scopeSelect),[r,a]=k.useState();if(Et(()=>{a(new DocumentFragment)},[]),!n.open){const o=r;return o?os.createPortal(b.jsx(Y6,{scope:e.__scopeSelect,children:b.jsx(Lf.Slot,{scope:e.__scopeSelect,children:b.jsx("div",{children:e.children})})}),o):null}return b.jsx(Q6,{...e,ref:t})});X6.displayName=go;var Cn=10,[Y6,ja]=us(go),bE="SelectContentImpl",PE=Al("SelectContent.RemoveScroll"),Q6=k.forwardRef((e,t)=>{const{__scopeSelect:n,position:r="item-aligned",onCloseAutoFocus:a,onEscapeKeyDown:o,onPointerDownOutside:i,side:s,sideOffset:l,align:u,alignOffset:p,arrowPadding:c,collisionBoundary:f,collisionPadding:m,sticky:h,hideWhenDetached:d,avoidCollisions:g,...v}=e,y=Ta(go,n),[x,P]=k.useState(null),[w,S]=k.useState(null),O=Xe(t,J=>P(J)),[C,_]=k.useState(null),[T,A]=k.useState(null),j=Ff(n),[$,M]=k.useState(!1),I=k.useRef(!1);k.useEffect(()=>{if(x)return TA(x)},[x]),t_();const R=k.useCallback(J=>{const[se,...q]=j().map(F=>F.ref.current),[K]=q.slice(-1),X=document.activeElement;for(const F of J)if(F===X||(F==null||F.scrollIntoView({block:"nearest"}),F===se&&w&&(w.scrollTop=0),F===K&&w&&(w.scrollTop=w.scrollHeight),F==null||F.focus(),document.activeElement!==X))return},[j,w]),L=k.useCallback(()=>R([C,x]),[R,C,x]);k.useEffect(()=>{$&&L()},[$,L]);const{onOpenChange:N,triggerPointerDownPosRef:D}=y;k.useEffect(()=>{if(x){let J={x:0,y:0};const se=K=>{var X,F;J={x:Math.abs(Math.round(K.pageX)-(((X=D.current)==null?void 0:X.x)??0)),y:Math.abs(Math.round(K.pageY)-(((F=D.current)==null?void 0:F.y)??0))}},q=K=>{J.x<=10&&J.y<=10?K.preventDefault():x.contains(K.target)||N(!1),document.removeEventListener("pointermove",se),D.current=null};return D.current!==null&&(document.addEventListener("pointermove",se),document.addEventListener("pointerup",q,{capture:!0,once:!0})),()=>{document.removeEventListener("pointermove",se),document.removeEventListener("pointerup",q,{capture:!0})}}},[x,N,D]),k.useEffect(()=>{const J=()=>N(!1);return window.addEventListener("blur",J),window.addEventListener("resize",J),()=>{window.removeEventListener("blur",J),window.removeEventListener("resize",J)}},[N]);const[H,W]=m4(J=>{const se=j().filter(X=>!X.disabled),q=se.find(X=>X.ref.current===document.activeElement),K=h4(se,J,q);K&&setTimeout(()=>K.ref.current.focus())}),G=k.useCallback((J,se,q)=>{const K=!I.current&&!q;(y.value!==void 0&&y.value===se||K)&&(_(J),K&&(I.current=!0))},[y.value]),Z=k.useCallback(()=>x==null?void 0:x.focus(),[x]),re=k.useCallback((J,se,q)=>{const K=!I.current&&!q;(y.value!==void 0&&y.value===se||K)&&A(J)},[y.value]),ve=r==="popper"?Gh:Z6,be=ve===Gh?{side:s,sideOffset:l,align:u,alignOffset:p,arrowPadding:c,collisionBoundary:f,collisionPadding:m,sticky:h,hideWhenDetached:d,avoidCollisions:g}:{};return b.jsx(Y6,{scope:n,content:x,viewport:w,onViewportChange:S,itemRefCallback:G,selectedItem:C,onItemLeave:Z,itemTextRefCallback:re,focusSelectedItem:L,selectedItemText:T,position:r,isPositioned:$,searchRef:H,children:b.jsx(z6,{as:PE,allowPinchZoom:!0,children:b.jsx(o6,{asChild:!0,trapped:y.open,onMountAutoFocus:J=>{J.preventDefault()},onUnmountAutoFocus:fe(a,J=>{var se;(se=y.trigger)==null||se.focus({preventScroll:!0}),J.preventDefault()}),children:b.jsx(wy,{asChild:!0,disableOutsidePointerEvents:!0,onEscapeKeyDown:o,onPointerDownOutside:i,onFocusOutside:J=>J.preventDefault(),onDismiss:()=>y.onOpenChange(!1),children:b.jsx(ve,{role:"listbox",id:y.contentId,"data-state":y.open?"open":"closed",dir:y.dir,onContextMenu:J=>J.preventDefault(),...v,...be,onPlaced:()=>M(!0),ref:O,style:{display:"flex",flexDirection:"column",outline:"none",...v.style},onKeyDown:fe(v.onKeyDown,J=>{const se=J.ctrlKey||J.altKey||J.metaKey;if(J.key==="Tab"&&J.preventDefault(),!se&&J.key.length===1&&W(J.key),["ArrowUp","ArrowDown","Home","End"].includes(J.key)){let K=j().filter(X=>!X.disabled).map(X=>X.ref.current);if(["ArrowUp","End"].includes(J.key)&&(K=K.slice().reverse()),["ArrowUp","ArrowDown"].includes(J.key)){const X=J.target,F=K.indexOf(X);K=K.slice(F+1)}setTimeout(()=>R(K)),J.preventDefault()}})})})})})})});Q6.displayName=bE;var SE="SelectItemAlignedPosition",Z6=k.forwardRef((e,t)=>{const{__scopeSelect:n,onPlaced:r,...a}=e,o=Ta(go,n),i=ja(go,n),[s,l]=k.useState(null),[u,p]=k.useState(null),c=Xe(t,O=>p(O)),f=Ff(n),m=k.useRef(!1),h=k.useRef(!0),{viewport:d,selectedItem:g,selectedItemText:v,focusSelectedItem:y}=i,x=k.useCallback(()=>{if(o.trigger&&o.valueNode&&s&&u&&d&&g&&v){const O=o.trigger.getBoundingClientRect(),C=u.getBoundingClientRect(),_=o.valueNode.getBoundingClientRect(),T=v.getBoundingClientRect();if(o.dir!=="rtl"){const X=T.left-C.left,F=_.left-X,pe=O.left-F,te=O.width+pe,$e=Math.max(te,C.width),Me=window.innerWidth-Cn,Qe=sx(F,[Cn,Math.max(Cn,Me-$e)]);s.style.minWidth=te+"px",s.style.left=Qe+"px"}else{const X=C.right-T.right,F=window.innerWidth-_.right-X,pe=window.innerWidth-O.right-F,te=O.width+pe,$e=Math.max(te,C.width),Me=window.innerWidth-Cn,Qe=sx(F,[Cn,Math.max(Cn,Me-$e)]);s.style.minWidth=te+"px",s.style.right=Qe+"px"}const A=f(),j=window.innerHeight-Cn*2,$=d.scrollHeight,M=window.getComputedStyle(u),I=parseInt(M.borderTopWidth,10),R=parseInt(M.paddingTop,10),L=parseInt(M.borderBottomWidth,10),N=parseInt(M.paddingBottom,10),D=I+R+$+N+L,H=Math.min(g.offsetHeight*5,D),W=window.getComputedStyle(d),G=parseInt(W.paddingTop,10),Z=parseInt(W.paddingBottom,10),re=O.top+O.height/2-Cn,ve=j-re,be=g.offsetHeight/2,J=g.offsetTop+be,se=I+R+J,q=D-se;if(se<=re){const X=A.length>0&&g===A[A.length-1].ref.current;s.style.bottom="0px";const F=u.clientHeight-d.offsetTop-d.offsetHeight,pe=Math.max(ve,be+(X?Z:0)+F+L),te=se+pe;s.style.height=te+"px"}else{const X=A.length>0&&g===A[0].ref.current;s.style.top="0px";const pe=Math.max(re,I+d.offsetTop+(X?G:0)+be)+q;s.style.height=pe+"px",d.scrollTop=se-re+d.offsetTop}s.style.margin=`${Cn}px 0`,s.style.minHeight=H+"px",s.style.maxHeight=j+"px",r==null||r(),requestAnimationFrame(()=>m.current=!0)}},[f,o.trigger,o.valueNode,s,u,d,g,v,o.dir,r]);Et(()=>x(),[x]);const[P,w]=k.useState();Et(()=>{u&&w(window.getComputedStyle(u).zIndex)},[u]);const S=k.useCallback(O=>{O&&h.current===!0&&(x(),y==null||y(),h.current=!1)},[x,y]);return b.jsx(kE,{scope:n,contentWrapper:s,shouldExpandOnScrollRef:m,onScrollButtonChange:S,children:b.jsx("div",{ref:l,style:{display:"flex",flexDirection:"column",position:"fixed",zIndex:P},children:b.jsx(Ae.div,{...a,ref:c,style:{boxSizing:"border-box",maxHeight:"100%",...a.style}})})})});Z6.displayName=SE;var OE="SelectPopperPosition",Gh=k.forwardRef((e,t)=>{const{__scopeSelect:n,align:r="start",collisionPadding:a=Cn,...o}=e,i=Bf(n);return b.jsx(A6,{...i,...o,ref:t,align:r,collisionPadding:a,style:{boxSizing:"border-box",...o.style,"--radix-select-content-transform-origin":"var(--radix-popper-transform-origin)","--radix-select-content-available-width":"var(--radix-popper-available-width)","--radix-select-content-available-height":"var(--radix-popper-available-height)","--radix-select-trigger-width":"var(--radix-popper-anchor-width)","--radix-select-trigger-height":"var(--radix-popper-anchor-height)"}})});Gh.displayName=OE;var[kE,Ty]=us(go,{}),Uh="SelectViewport",J6=k.forwardRef((e,t)=>{const{__scopeSelect:n,nonce:r,...a}=e,o=ja(Uh,n),i=Ty(Uh,n),s=Xe(t,o.onViewportChange),l=k.useRef(0);return b.jsxs(b.Fragment,{children:[b.jsx("style",{dangerouslySetInnerHTML:{__html:"[data-radix-select-viewport]{scrollbar-width:none;-ms-overflow-style:none;-webkit-overflow-scrolling:touch;}[data-radix-select-viewport]::-webkit-scrollbar{display:none}"},nonce:r}),b.jsx(Lf.Slot,{scope:n,children:b.jsx(Ae.div,{"data-radix-select-viewport":"",role:"presentation",...a,ref:s,style:{position:"relative",flex:1,overflow:"hidden auto",...a.style},onScroll:fe(a.onScroll,u=>{const p=u.currentTarget,{contentWrapper:c,shouldExpandOnScrollRef:f}=i;if(f!=null&&f.current&&c){const m=Math.abs(l.current-p.scrollTop);if(m>0){const h=window.innerHeight-Cn*2,d=parseFloat(c.style.minHeight),g=parseFloat(c.style.height),v=Math.max(d,g);if(v<h){const y=v+m,x=Math.min(h,y),P=y-x;c.style.height=x+"px",c.style.bottom==="0px"&&(p.scrollTop=P>0?P:0,c.style.justifyContent="flex-end")}}}l.current=p.scrollTop})})})]})});J6.displayName=Uh;var e4="SelectGroup",[CE,_E]=us(e4),AE=k.forwardRef((e,t)=>{const{__scopeSelect:n,...r}=e,a=Pu();return b.jsx(CE,{scope:n,id:a,children:b.jsx(Ae.div,{role:"group","aria-labelledby":a,...r,ref:t})})});AE.displayName=e4;var t4="SelectLabel",n4=k.forwardRef((e,t)=>{const{__scopeSelect:n,...r}=e,a=_E(t4,n);return b.jsx(Ae.div,{id:a.id,...r,ref:t})});n4.displayName=t4;var fp="SelectItem",[EE,r4]=us(fp),a4=k.forwardRef((e,t)=>{const{__scopeSelect:n,value:r,disabled:a=!1,textValue:o,...i}=e,s=Ta(fp,n),l=ja(fp,n),u=s.value===r,[p,c]=k.useState(o??""),[f,m]=k.useState(!1),h=Xe(t,y=>{var x;return(x=l.itemRefCallback)==null?void 0:x.call(l,y,r,a)}),d=Pu(),g=k.useRef("touch"),v=()=>{a||(s.onValueChange(r),s.onOpenChange(!1))};if(r==="")throw new Error("A <Select.Item /> must have a value prop that is not an empty string. This is because the Select value can be set to an empty string to clear the selection and show the placeholder.");return b.jsx(EE,{scope:n,value:r,disabled:a,textId:d,isSelected:u,onItemTextChange:k.useCallback(y=>{c(x=>x||((y==null?void 0:y.textContent)??"").trim())},[]),children:b.jsx(Lf.ItemSlot,{scope:n,value:r,disabled:a,textValue:p,children:b.jsx(Ae.div,{role:"option","aria-labelledby":d,"data-highlighted":f?"":void 0,"aria-selected":u&&f,"data-state":u?"checked":"unchecked","aria-disabled":a||void 0,"data-disabled":a?"":void 0,tabIndex:a?void 0:-1,...i,ref:h,onFocus:fe(i.onFocus,()=>m(!0)),onBlur:fe(i.onBlur,()=>m(!1)),onClick:fe(i.onClick,()=>{g.current!=="mouse"&&v()}),onPointerUp:fe(i.onPointerUp,()=>{g.current==="mouse"&&v()}),onPointerDown:fe(i.onPointerDown,y=>{g.current=y.pointerType}),onPointerMove:fe(i.onPointerMove,y=>{var x;g.current=y.pointerType,a?(x=l.onItemLeave)==null||x.call(l):g.current==="mouse"&&y.currentTarget.focus({preventScroll:!0})}),onPointerLeave:fe(i.onPointerLeave,y=>{var x;y.currentTarget===document.activeElement&&((x=l.onItemLeave)==null||x.call(l))}),onKeyDown:fe(i.onKeyDown,y=>{var P;((P=l.searchRef)==null?void 0:P.current)!==""&&y.key===" "||(mE.includes(y.key)&&v(),y.key===" "&&y.preventDefault())})})})})});a4.displayName=fp;var Xs="SelectItemText",o4=k.forwardRef((e,t)=>{const{__scopeSelect:n,className:r,style:a,...o}=e,i=Ta(Xs,n),s=ja(Xs,n),l=r4(Xs,n),u=gE(Xs,n),[p,c]=k.useState(null),f=Xe(t,v=>c(v),l.onItemTextChange,v=>{var y;return(y=s.itemTextRefCallback)==null?void 0:y.call(s,v,l.value,l.disabled)}),m=p==null?void 0:p.textContent,h=k.useMemo(()=>b.jsx("option",{value:l.value,disabled:l.disabled,children:m},l.value),[l.disabled,l.value,m]),{onNativeOptionAdd:d,onNativeOptionRemove:g}=u;return Et(()=>(d(h),()=>g(h)),[d,g,h]),b.jsxs(b.Fragment,{children:[b.jsx(Ae.span,{id:l.textId,...o,ref:f}),l.isSelected&&i.valueNode&&!i.valueNodeHasChildren?os.createPortal(o.children,i.valueNode):null]})});o4.displayName=Xs;var i4="SelectItemIndicator",s4=k.forwardRef((e,t)=>{const{__scopeSelect:n,...r}=e;return r4(i4,n).isSelected?b.jsx(Ae.span,{"aria-hidden":!0,...r,ref:t}):null});s4.displayName=i4;var Wh="SelectScrollUpButton",l4=k.forwardRef((e,t)=>{const n=ja(Wh,e.__scopeSelect),r=Ty(Wh,e.__scopeSelect),[a,o]=k.useState(!1),i=Xe(t,r.onScrollButtonChange);return Et(()=>{if(n.viewport&&n.isPositioned){let s=function(){const u=l.scrollTop>0;o(u)};const l=n.viewport;return s(),l.addEventListener("scroll",s),()=>l.removeEventListener("scroll",s)}},[n.viewport,n.isPositioned]),a?b.jsx(c4,{...e,ref:i,onAutoScroll:()=>{const{viewport:s,selectedItem:l}=n;s&&l&&(s.scrollTop=s.scrollTop-l.offsetHeight)}}):null});l4.displayName=Wh;var qh="SelectScrollDownButton",u4=k.forwardRef((e,t)=>{const n=ja(qh,e.__scopeSelect),r=Ty(qh,e.__scopeSelect),[a,o]=k.useState(!1),i=Xe(t,r.onScrollButtonChange);return Et(()=>{if(n.viewport&&n.isPositioned){let s=function(){const u=l.scrollHeight-l.clientHeight,p=Math.ceil(l.scrollTop)<u;o(p)};const l=n.viewport;return s(),l.addEventListener("scroll",s),()=>l.removeEventListener("scroll",s)}},[n.viewport,n.isPositioned]),a?b.jsx(c4,{...e,ref:i,onAutoScroll:()=>{const{viewport:s,selectedItem:l}=n;s&&l&&(s.scrollTop=s.scrollTop+l.offsetHeight)}}):null});u4.displayName=qh;var c4=k.forwardRef((e,t)=>{const{__scopeSelect:n,onAutoScroll:r,...a}=e,o=ja("SelectScrollButton",n),i=k.useRef(null),s=Ff(n),l=k.useCallback(()=>{i.current!==null&&(window.clearInterval(i.current),i.current=null)},[]);return k.useEffect(()=>()=>l(),[l]),Et(()=>{var p;const u=s().find(c=>c.ref.current===document.activeElement);(p=u==null?void 0:u.ref.current)==null||p.scrollIntoView({block:"nearest"})},[s]),b.jsx(Ae.div,{"aria-hidden":!0,...a,ref:t,style:{flexShrink:0,...a.style},onPointerDown:fe(a.onPointerDown,()=>{i.current===null&&(i.current=window.setInterval(r,50))}),onPointerMove:fe(a.onPointerMove,()=>{var u;(u=o.onItemLeave)==null||u.call(o),i.current===null&&(i.current=window.setInterval(r,50))}),onPointerLeave:fe(a.onPointerLeave,()=>{l()})})}),TE="SelectSeparator",p4=k.forwardRef((e,t)=>{const{__scopeSelect:n,...r}=e;return b.jsx(Ae.div,{"aria-hidden":!0,...r,ref:t})});p4.displayName=TE;var Vh="SelectArrow",jE=k.forwardRef((e,t)=>{const{__scopeSelect:n,...r}=e,a=Bf(n),o=Ta(Vh,n),i=ja(Vh,n);return o.open&&i.position==="popper"?b.jsx(E6,{...a,...r,ref:t}):null});jE.displayName=Vh;var NE="SelectBubbleInput",f4=k.forwardRef(({__scopeSelect:e,value:t,...n},r)=>{const a=k.useRef(null),o=Xe(r,a),i=OA(t);return k.useEffect(()=>{const s=a.current;if(!s)return;const l=window.HTMLSelectElement.prototype,p=Object.getOwnPropertyDescriptor(l,"value").set;if(i!==t&&p){const c=new Event("change",{bubbles:!0});p.call(s,t),s.dispatchEvent(c)}},[i,t]),b.jsx(Ae.select,{...n,style:{...j6,...n.style},ref:o,defaultValue:t})});f4.displayName=NE;function d4(e){return e===""||e===void 0}function m4(e){const t=Oa(e),n=k.useRef(""),r=k.useRef(0),a=k.useCallback(i=>{const s=n.current+i;t(s),function l(u){n.current=u,window.clearTimeout(r.current),u!==""&&(r.current=window.setTimeout(()=>l(""),1e3))}(s)},[t]),o=k.useCallback(()=>{n.current="",window.clearTimeout(r.current)},[]);return k.useEffect(()=>()=>window.clearTimeout(r.current),[]),[n,a,o]}function h4(e,t,n){const a=t.length>1&&Array.from(t).every(u=>u===t[0])?t[0]:t,o=n?e.indexOf(n):-1;let i=$E(e,Math.max(o,0));a.length===1&&(i=i.filter(u=>u!==n));const l=i.find(u=>u.textValue.toLowerCase().startsWith(a.toLowerCase()));return l!==n?l:void 0}function $E(e,t){return e.map((n,r)=>e[(t+r)%e.length])}var ME=H6,v4=U6,RE=q6,IE=V6,DE=K6,y4=X6,LE=J6,g4=n4,x4=a4,FE=o4,BE=s4,w4=l4,b4=u4,P4=p4;/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const zE=e=>e.replace(/([a-z0-9])([A-Z])/g,"$1-$2").toLowerCase(),S4=(...e)=>e.filter((t,n,r)=>!!t&&t.trim()!==""&&r.indexOf(t)===n).join(" ").trim();/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var HE={xmlns:"http://www.w3.org/2000/svg",width:24,height:24,viewBox:"0 0 24 24",fill:"none",stroke:"currentColor",strokeWidth:2,strokeLinecap:"round",strokeLinejoin:"round"};/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const GE=k.forwardRef(({color:e="currentColor",size:t=24,strokeWidth:n=2,absoluteStrokeWidth:r,className:a="",children:o,iconNode:i,...s},l)=>k.createElement("svg",{ref:l,...HE,width:t,height:t,stroke:e,strokeWidth:r?Number(n)*24/Number(t):n,className:S4("lucide",a),...s},[...i.map(([u,p])=>k.createElement(u,p)),...Array.isArray(o)?o:[o]]));/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const Ye=(e,t)=>{const n=k.forwardRef(({className:r,...a},o)=>k.createElement(GE,{ref:o,iconNode:t,className:S4(`lucide-${zE(e)}`,r),...a}));return n.displayName=`${e}`,n};/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const O4=Ye("Activity",[["path",{d:"M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2",key:"169zse"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const UE=Ye("Calendar",[["path",{d:"M8 2v4",key:"1cmpym"}],["path",{d:"M16 2v4",key:"4m81vk"}],["rect",{width:"18",height:"18",x:"3",y:"4",rx:"2",key:"1hopcy"}],["path",{d:"M3 10h18",key:"8toen8"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const k4=Ye("ChartColumn",[["path",{d:"M3 3v16a2 2 0 0 0 2 2h16",key:"c24i48"}],["path",{d:"M18 17V9",key:"2bz60n"}],["path",{d:"M13 17V5",key:"1frdt8"}],["path",{d:"M8 17v-3",key:"17ska0"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const WE=Ye("Check",[["path",{d:"M20 6 9 17l-5-5",key:"1gmf2c"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const C4=Ye("ChevronDown",[["path",{d:"m6 9 6 6 6-6",key:"qrunsl"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const qE=Ye("ChevronUp",[["path",{d:"m18 15-6-6-6 6",key:"153udz"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const VE=Ye("CircleAlert",[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["line",{x1:"12",x2:"12",y1:"8",y2:"12",key:"1pkeuh"}],["line",{x1:"12",x2:"12.01",y1:"16",y2:"16",key:"4dfq90"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const _4=Ye("CircleCheckBig",[["path",{d:"M21.801 10A10 10 0 1 1 17 3.335",key:"yps3ct"}],["path",{d:"m9 11 3 3L22 4",key:"1pflzl"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const A4=Ye("Clock",[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["polyline",{points:"12 6 12 12 16 14",key:"68esgv"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const E4=Ye("Database",[["ellipse",{cx:"12",cy:"5",rx:"9",ry:"3",key:"msslwz"}],["path",{d:"M3 5V19A9 3 0 0 0 21 19V5",key:"1wlel7"}],["path",{d:"M3 12A9 3 0 0 0 21 12",key:"mv7ke4"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const KE=Ye("ExternalLink",[["path",{d:"M15 3h6v6",key:"1q9fwt"}],["path",{d:"M10 14 21 3",key:"gplh6r"}],["path",{d:"M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6",key:"a6xqqp"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const XE=Ye("Gauge",[["path",{d:"m12 14 4-4",key:"9kzdfg"}],["path",{d:"M3.34 19a10 10 0 1 1 17.32 0",key:"19p75a"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const T4=Ye("GitBranch",[["line",{x1:"6",x2:"6",y1:"3",y2:"15",key:"17qcm7"}],["circle",{cx:"18",cy:"6",r:"3",key:"1h7g24"}],["circle",{cx:"6",cy:"18",r:"3",key:"fqmcym"}],["path",{d:"M18 9a9 9 0 0 1-9 9",key:"n2h4wq"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const YE=Ye("MemoryStick",[["path",{d:"M6 19v-3",key:"1nvgqn"}],["path",{d:"M10 19v-3",key:"iu8nkm"}],["path",{d:"M14 19v-3",key:"kcehxu"}],["path",{d:"M18 19v-3",key:"1vh91z"}],["path",{d:"M8 11V9",key:"63erz4"}],["path",{d:"M16 11V9",key:"fru6f3"}],["path",{d:"M12 11V9",key:"ha00sb"}],["path",{d:"M2 15h20",key:"16ne18"}],["path",{d:"M2 7a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v1.1a2 2 0 0 0 0 3.837V17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-5.1a2 2 0 0 0 0-3.837Z",key:"lhddv3"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const j4=Ye("RefreshCw",[["path",{d:"M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8",key:"v9h5vc"}],["path",{d:"M21 3v5h-5",key:"1q7to0"}],["path",{d:"M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16",key:"3uifl3"}],["path",{d:"M8 16H3v5",key:"1cv678"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const N4=Ye("Server",[["rect",{width:"20",height:"8",x:"2",y:"2",rx:"2",ry:"2",key:"ngkwjq"}],["rect",{width:"20",height:"8",x:"2",y:"14",rx:"2",ry:"2",key:"iecqi9"}],["line",{x1:"6",x2:"6.01",y1:"6",y2:"6",key:"16zg32"}],["line",{x1:"6",x2:"6.01",y1:"18",y2:"18",key:"nzw8ys"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const QE=Ye("Shield",[["path",{d:"M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z",key:"oel41y"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const ZE=Ye("Target",[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["circle",{cx:"12",cy:"12",r:"6",key:"1vlfrh"}],["circle",{cx:"12",cy:"12",r:"2",key:"1c9p78"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const $4=Ye("TrendingUp",[["polyline",{points:"22 7 13.5 15.5 8.5 10.5 2 17",key:"126l90"}],["polyline",{points:"16 7 22 7 22 13",key:"kwv8wd"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const JE=Ye("TriangleAlert",[["path",{d:"m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3",key:"wmoenq"}],["path",{d:"M12 9v4",key:"juzpu7"}],["path",{d:"M12 17h.01",key:"p32p05"}]]);/**
 * @license lucide-react v0.462.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const eT=Ye("Zap",[["path",{d:"M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z",key:"1xq2db"}]]);function M4(e){var t,n,r="";if(typeof e=="string"||typeof e=="number")r+=e;else if(typeof e=="object")if(Array.isArray(e)){var a=e.length;for(t=0;t<a;t++)e[t]&&(n=M4(e[t]))&&(r&&(r+=" "),r+=n)}else for(n in e)e[n]&&(r&&(r+=" "),r+=n);return r}function ue(){for(var e,t,n=0,r="",a=arguments.length;n<a;n++)(e=arguments[n])&&(t=M4(e))&&(r&&(r+=" "),r+=t);return r}const jy="-",tT=e=>{const t=rT(e),{conflictingClassGroups:n,conflictingClassGroupModifiers:r}=e;return{getClassGroupId:i=>{const s=i.split(jy);return s[0]===""&&s.length!==1&&s.shift(),R4(s,t)||nT(i)},getConflictingClassGroupIds:(i,s)=>{const l=n[i]||[];return s&&r[i]?[...l,...r[i]]:l}}},R4=(e,t)=>{var i;if(e.length===0)return t.classGroupId;const n=e[0],r=t.nextPart.get(n),a=r?R4(e.slice(1),r):void 0;if(a)return a;if(t.validators.length===0)return;const o=e.join(jy);return(i=t.validators.find(({validator:s})=>s(o)))==null?void 0:i.classGroupId},Ex=/^\[(.+)\]$/,nT=e=>{if(Ex.test(e)){const t=Ex.exec(e)[1],n=t==null?void 0:t.substring(0,t.indexOf(":"));if(n)return"arbitrary.."+n}},rT=e=>{const{theme:t,prefix:n}=e,r={nextPart:new Map,validators:[]};return oT(Object.entries(e.classGroups),n).forEach(([o,i])=>{Kh(i,r,o,t)}),r},Kh=(e,t,n,r)=>{e.forEach(a=>{if(typeof a=="string"){const o=a===""?t:Tx(t,a);o.classGroupId=n;return}if(typeof a=="function"){if(aT(a)){Kh(a(r),t,n,r);return}t.validators.push({validator:a,classGroupId:n});return}Object.entries(a).forEach(([o,i])=>{Kh(i,Tx(t,o),n,r)})})},Tx=(e,t)=>{let n=e;return t.split(jy).forEach(r=>{n.nextPart.has(r)||n.nextPart.set(r,{nextPart:new Map,validators:[]}),n=n.nextPart.get(r)}),n},aT=e=>e.isThemeGetter,oT=(e,t)=>t?e.map(([n,r])=>{const a=r.map(o=>typeof o=="string"?t+o:typeof o=="object"?Object.fromEntries(Object.entries(o).map(([i,s])=>[t+i,s])):o);return[n,a]}):e,iT=e=>{if(e<1)return{get:()=>{},set:()=>{}};let t=0,n=new Map,r=new Map;const a=(o,i)=>{n.set(o,i),t++,t>e&&(t=0,r=n,n=new Map)};return{get(o){let i=n.get(o);if(i!==void 0)return i;if((i=r.get(o))!==void 0)return a(o,i),i},set(o,i){n.has(o)?n.set(o,i):a(o,i)}}},I4="!",sT=e=>{const{separator:t,experimentalParseClassName:n}=e,r=t.length===1,a=t[0],o=t.length,i=s=>{const l=[];let u=0,p=0,c;for(let g=0;g<s.length;g++){let v=s[g];if(u===0){if(v===a&&(r||s.slice(g,g+o)===t)){l.push(s.slice(p,g)),p=g+o;continue}if(v==="/"){c=g;continue}}v==="["?u++:v==="]"&&u--}const f=l.length===0?s:s.substring(p),m=f.startsWith(I4),h=m?f.substring(1):f,d=c&&c>p?c-p:void 0;return{modifiers:l,hasImportantModifier:m,baseClassName:h,maybePostfixModifierPosition:d}};return n?s=>n({className:s,parseClassName:i}):i},lT=e=>{if(e.length<=1)return e;const t=[];let n=[];return e.forEach(r=>{r[0]==="["?(t.push(...n.sort(),r),n=[]):n.push(r)}),t.push(...n.sort()),t},uT=e=>({cache:iT(e.cacheSize),parseClassName:sT(e),...tT(e)}),cT=/\s+/,pT=(e,t)=>{const{parseClassName:n,getClassGroupId:r,getConflictingClassGroupIds:a}=t,o=[],i=e.trim().split(cT);let s="";for(let l=i.length-1;l>=0;l-=1){const u=i[l],{modifiers:p,hasImportantModifier:c,baseClassName:f,maybePostfixModifierPosition:m}=n(u);let h=!!m,d=r(h?f.substring(0,m):f);if(!d){if(!h){s=u+(s.length>0?" "+s:s);continue}if(d=r(f),!d){s=u+(s.length>0?" "+s:s);continue}h=!1}const g=lT(p).join(":"),v=c?g+I4:g,y=v+d;if(o.includes(y))continue;o.push(y);const x=a(d,h);for(let P=0;P<x.length;++P){const w=x[P];o.push(v+w)}s=u+(s.length>0?" "+s:s)}return s};function fT(){let e=0,t,n,r="";for(;e<arguments.length;)(t=arguments[e++])&&(n=D4(t))&&(r&&(r+=" "),r+=n);return r}const D4=e=>{if(typeof e=="string")return e;let t,n="";for(let r=0;r<e.length;r++)e[r]&&(t=D4(e[r]))&&(n&&(n+=" "),n+=t);return n};function dT(e,...t){let n,r,a,o=i;function i(l){const u=t.reduce((p,c)=>c(p),e());return n=uT(u),r=n.cache.get,a=n.cache.set,o=s,s(l)}function s(l){const u=r(l);if(u)return u;const p=pT(l,n);return a(l,p),p}return function(){return o(fT.apply(null,arguments))}}const Re=e=>{const t=n=>n[e]||[];return t.isThemeGetter=!0,t},L4=/^\[(?:([a-z-]+):)?(.+)\]$/i,mT=/^\d+\/\d+$/,hT=new Set(["px","full","screen"]),vT=/^(\d+(\.\d+)?)?(xs|sm|md|lg|xl)$/,yT=/\d+(%|px|r?em|[sdl]?v([hwib]|min|max)|pt|pc|in|cm|mm|cap|ch|ex|r?lh|cq(w|h|i|b|min|max))|\b(calc|min|max|clamp)\(.+\)|^0$/,gT=/^(rgba?|hsla?|hwb|(ok)?(lab|lch))\(.+\)$/,xT=/^(inset_)?-?((\d+)?\.?(\d+)[a-z]+|0)_-?((\d+)?\.?(\d+)[a-z]+|0)/,wT=/^(url|image|image-set|cross-fade|element|(repeating-)?(linear|radial|conic)-gradient)\(.+\)$/,yr=e=>li(e)||hT.has(e)||mT.test(e),Vr=e=>cs(e,"length",AT),li=e=>!!e&&!Number.isNaN(Number(e)),hm=e=>cs(e,"number",li),$s=e=>!!e&&Number.isInteger(Number(e)),bT=e=>e.endsWith("%")&&li(e.slice(0,-1)),oe=e=>L4.test(e),Kr=e=>vT.test(e),PT=new Set(["length","size","percentage"]),ST=e=>cs(e,PT,F4),OT=e=>cs(e,"position",F4),kT=new Set(["image","url"]),CT=e=>cs(e,kT,TT),_T=e=>cs(e,"",ET),Ms=()=>!0,cs=(e,t,n)=>{const r=L4.exec(e);return r?r[1]?typeof t=="string"?r[1]===t:t.has(r[1]):n(r[2]):!1},AT=e=>yT.test(e)&&!gT.test(e),F4=()=>!1,ET=e=>xT.test(e),TT=e=>wT.test(e),jT=()=>{const e=Re("colors"),t=Re("spacing"),n=Re("blur"),r=Re("brightness"),a=Re("borderColor"),o=Re("borderRadius"),i=Re("borderSpacing"),s=Re("borderWidth"),l=Re("contrast"),u=Re("grayscale"),p=Re("hueRotate"),c=Re("invert"),f=Re("gap"),m=Re("gradientColorStops"),h=Re("gradientColorStopPositions"),d=Re("inset"),g=Re("margin"),v=Re("opacity"),y=Re("padding"),x=Re("saturate"),P=Re("scale"),w=Re("sepia"),S=Re("skew"),O=Re("space"),C=Re("translate"),_=()=>["auto","contain","none"],T=()=>["auto","hidden","clip","visible","scroll"],A=()=>["auto",oe,t],j=()=>[oe,t],$=()=>["",yr,Vr],M=()=>["auto",li,oe],I=()=>["bottom","center","left","left-bottom","left-top","right","right-bottom","right-top","top"],R=()=>["solid","dashed","dotted","double","none"],L=()=>["normal","multiply","screen","overlay","darken","lighten","color-dodge","color-burn","hard-light","soft-light","difference","exclusion","hue","saturation","color","luminosity"],N=()=>["start","end","center","between","around","evenly","stretch"],D=()=>["","0",oe],H=()=>["auto","avoid","all","avoid-page","page","left","right","column"],W=()=>[li,oe];return{cacheSize:500,separator:":",theme:{colors:[Ms],spacing:[yr,Vr],blur:["none","",Kr,oe],brightness:W(),borderColor:[e],borderRadius:["none","","full",Kr,oe],borderSpacing:j(),borderWidth:$(),contrast:W(),grayscale:D(),hueRotate:W(),invert:D(),gap:j(),gradientColorStops:[e],gradientColorStopPositions:[bT,Vr],inset:A(),margin:A(),opacity:W(),padding:j(),saturate:W(),scale:W(),sepia:D(),skew:W(),space:j(),translate:j()},classGroups:{aspect:[{aspect:["auto","square","video",oe]}],container:["container"],columns:[{columns:[Kr]}],"break-after":[{"break-after":H()}],"break-before":[{"break-before":H()}],"break-inside":[{"break-inside":["auto","avoid","avoid-page","avoid-column"]}],"box-decoration":[{"box-decoration":["slice","clone"]}],box:[{box:["border","content"]}],display:["block","inline-block","inline","flex","inline-flex","table","inline-table","table-caption","table-cell","table-column","table-column-group","table-footer-group","table-header-group","table-row-group","table-row","flow-root","grid","inline-grid","contents","list-item","hidden"],float:[{float:["right","left","none","start","end"]}],clear:[{clear:["left","right","both","none","start","end"]}],isolation:["isolate","isolation-auto"],"object-fit":[{object:["contain","cover","fill","none","scale-down"]}],"object-position":[{object:[...I(),oe]}],overflow:[{overflow:T()}],"overflow-x":[{"overflow-x":T()}],"overflow-y":[{"overflow-y":T()}],overscroll:[{overscroll:_()}],"overscroll-x":[{"overscroll-x":_()}],"overscroll-y":[{"overscroll-y":_()}],position:["static","fixed","absolute","relative","sticky"],inset:[{inset:[d]}],"inset-x":[{"inset-x":[d]}],"inset-y":[{"inset-y":[d]}],start:[{start:[d]}],end:[{end:[d]}],top:[{top:[d]}],right:[{right:[d]}],bottom:[{bottom:[d]}],left:[{left:[d]}],visibility:["visible","invisible","collapse"],z:[{z:["auto",$s,oe]}],basis:[{basis:A()}],"flex-direction":[{flex:["row","row-reverse","col","col-reverse"]}],"flex-wrap":[{flex:["wrap","wrap-reverse","nowrap"]}],flex:[{flex:["1","auto","initial","none",oe]}],grow:[{grow:D()}],shrink:[{shrink:D()}],order:[{order:["first","last","none",$s,oe]}],"grid-cols":[{"grid-cols":[Ms]}],"col-start-end":[{col:["auto",{span:["full",$s,oe]},oe]}],"col-start":[{"col-start":M()}],"col-end":[{"col-end":M()}],"grid-rows":[{"grid-rows":[Ms]}],"row-start-end":[{row:["auto",{span:[$s,oe]},oe]}],"row-start":[{"row-start":M()}],"row-end":[{"row-end":M()}],"grid-flow":[{"grid-flow":["row","col","dense","row-dense","col-dense"]}],"auto-cols":[{"auto-cols":["auto","min","max","fr",oe]}],"auto-rows":[{"auto-rows":["auto","min","max","fr",oe]}],gap:[{gap:[f]}],"gap-x":[{"gap-x":[f]}],"gap-y":[{"gap-y":[f]}],"justify-content":[{justify:["normal",...N()]}],"justify-items":[{"justify-items":["start","end","center","stretch"]}],"justify-self":[{"justify-self":["auto","start","end","center","stretch"]}],"align-content":[{content:["normal",...N(),"baseline"]}],"align-items":[{items:["start","end","center","baseline","stretch"]}],"align-self":[{self:["auto","start","end","center","stretch","baseline"]}],"place-content":[{"place-content":[...N(),"baseline"]}],"place-items":[{"place-items":["start","end","center","baseline","stretch"]}],"place-self":[{"place-self":["auto","start","end","center","stretch"]}],p:[{p:[y]}],px:[{px:[y]}],py:[{py:[y]}],ps:[{ps:[y]}],pe:[{pe:[y]}],pt:[{pt:[y]}],pr:[{pr:[y]}],pb:[{pb:[y]}],pl:[{pl:[y]}],m:[{m:[g]}],mx:[{mx:[g]}],my:[{my:[g]}],ms:[{ms:[g]}],me:[{me:[g]}],mt:[{mt:[g]}],mr:[{mr:[g]}],mb:[{mb:[g]}],ml:[{ml:[g]}],"space-x":[{"space-x":[O]}],"space-x-reverse":["space-x-reverse"],"space-y":[{"space-y":[O]}],"space-y-reverse":["space-y-reverse"],w:[{w:["auto","min","max","fit","svw","lvw","dvw",oe,t]}],"min-w":[{"min-w":[oe,t,"min","max","fit"]}],"max-w":[{"max-w":[oe,t,"none","full","min","max","fit","prose",{screen:[Kr]},Kr]}],h:[{h:[oe,t,"auto","min","max","fit","svh","lvh","dvh"]}],"min-h":[{"min-h":[oe,t,"min","max","fit","svh","lvh","dvh"]}],"max-h":[{"max-h":[oe,t,"min","max","fit","svh","lvh","dvh"]}],size:[{size:[oe,t,"auto","min","max","fit"]}],"font-size":[{text:["base",Kr,Vr]}],"font-smoothing":["antialiased","subpixel-antialiased"],"font-style":["italic","not-italic"],"font-weight":[{font:["thin","extralight","light","normal","medium","semibold","bold","extrabold","black",hm]}],"font-family":[{font:[Ms]}],"fvn-normal":["normal-nums"],"fvn-ordinal":["ordinal"],"fvn-slashed-zero":["slashed-zero"],"fvn-figure":["lining-nums","oldstyle-nums"],"fvn-spacing":["proportional-nums","tabular-nums"],"fvn-fraction":["diagonal-fractions","stacked-fractions"],tracking:[{tracking:["tighter","tight","normal","wide","wider","widest",oe]}],"line-clamp":[{"line-clamp":["none",li,hm]}],leading:[{leading:["none","tight","snug","normal","relaxed","loose",yr,oe]}],"list-image":[{"list-image":["none",oe]}],"list-style-type":[{list:["none","disc","decimal",oe]}],"list-style-position":[{list:["inside","outside"]}],"placeholder-color":[{placeholder:[e]}],"placeholder-opacity":[{"placeholder-opacity":[v]}],"text-alignment":[{text:["left","center","right","justify","start","end"]}],"text-color":[{text:[e]}],"text-opacity":[{"text-opacity":[v]}],"text-decoration":["underline","overline","line-through","no-underline"],"text-decoration-style":[{decoration:[...R(),"wavy"]}],"text-decoration-thickness":[{decoration:["auto","from-font",yr,Vr]}],"underline-offset":[{"underline-offset":["auto",yr,oe]}],"text-decoration-color":[{decoration:[e]}],"text-transform":["uppercase","lowercase","capitalize","normal-case"],"text-overflow":["truncate","text-ellipsis","text-clip"],"text-wrap":[{text:["wrap","nowrap","balance","pretty"]}],indent:[{indent:j()}],"vertical-align":[{align:["baseline","top","middle","bottom","text-top","text-bottom","sub","super",oe]}],whitespace:[{whitespace:["normal","nowrap","pre","pre-line","pre-wrap","break-spaces"]}],break:[{break:["normal","words","all","keep"]}],hyphens:[{hyphens:["none","manual","auto"]}],content:[{content:["none",oe]}],"bg-attachment":[{bg:["fixed","local","scroll"]}],"bg-clip":[{"bg-clip":["border","padding","content","text"]}],"bg-opacity":[{"bg-opacity":[v]}],"bg-origin":[{"bg-origin":["border","padding","content"]}],"bg-position":[{bg:[...I(),OT]}],"bg-repeat":[{bg:["no-repeat",{repeat:["","x","y","round","space"]}]}],"bg-size":[{bg:["auto","cover","contain",ST]}],"bg-image":[{bg:["none",{"gradient-to":["t","tr","r","br","b","bl","l","tl"]},CT]}],"bg-color":[{bg:[e]}],"gradient-from-pos":[{from:[h]}],"gradient-via-pos":[{via:[h]}],"gradient-to-pos":[{to:[h]}],"gradient-from":[{from:[m]}],"gradient-via":[{via:[m]}],"gradient-to":[{to:[m]}],rounded:[{rounded:[o]}],"rounded-s":[{"rounded-s":[o]}],"rounded-e":[{"rounded-e":[o]}],"rounded-t":[{"rounded-t":[o]}],"rounded-r":[{"rounded-r":[o]}],"rounded-b":[{"rounded-b":[o]}],"rounded-l":[{"rounded-l":[o]}],"rounded-ss":[{"rounded-ss":[o]}],"rounded-se":[{"rounded-se":[o]}],"rounded-ee":[{"rounded-ee":[o]}],"rounded-es":[{"rounded-es":[o]}],"rounded-tl":[{"rounded-tl":[o]}],"rounded-tr":[{"rounded-tr":[o]}],"rounded-br":[{"rounded-br":[o]}],"rounded-bl":[{"rounded-bl":[o]}],"border-w":[{border:[s]}],"border-w-x":[{"border-x":[s]}],"border-w-y":[{"border-y":[s]}],"border-w-s":[{"border-s":[s]}],"border-w-e":[{"border-e":[s]}],"border-w-t":[{"border-t":[s]}],"border-w-r":[{"border-r":[s]}],"border-w-b":[{"border-b":[s]}],"border-w-l":[{"border-l":[s]}],"border-opacity":[{"border-opacity":[v]}],"border-style":[{border:[...R(),"hidden"]}],"divide-x":[{"divide-x":[s]}],"divide-x-reverse":["divide-x-reverse"],"divide-y":[{"divide-y":[s]}],"divide-y-reverse":["divide-y-reverse"],"divide-opacity":[{"divide-opacity":[v]}],"divide-style":[{divide:R()}],"border-color":[{border:[a]}],"border-color-x":[{"border-x":[a]}],"border-color-y":[{"border-y":[a]}],"border-color-s":[{"border-s":[a]}],"border-color-e":[{"border-e":[a]}],"border-color-t":[{"border-t":[a]}],"border-color-r":[{"border-r":[a]}],"border-color-b":[{"border-b":[a]}],"border-color-l":[{"border-l":[a]}],"divide-color":[{divide:[a]}],"outline-style":[{outline:["",...R()]}],"outline-offset":[{"outline-offset":[yr,oe]}],"outline-w":[{outline:[yr,Vr]}],"outline-color":[{outline:[e]}],"ring-w":[{ring:$()}],"ring-w-inset":["ring-inset"],"ring-color":[{ring:[e]}],"ring-opacity":[{"ring-opacity":[v]}],"ring-offset-w":[{"ring-offset":[yr,Vr]}],"ring-offset-color":[{"ring-offset":[e]}],shadow:[{shadow:["","inner","none",Kr,_T]}],"shadow-color":[{shadow:[Ms]}],opacity:[{opacity:[v]}],"mix-blend":[{"mix-blend":[...L(),"plus-lighter","plus-darker"]}],"bg-blend":[{"bg-blend":L()}],filter:[{filter:["","none"]}],blur:[{blur:[n]}],brightness:[{brightness:[r]}],contrast:[{contrast:[l]}],"drop-shadow":[{"drop-shadow":["","none",Kr,oe]}],grayscale:[{grayscale:[u]}],"hue-rotate":[{"hue-rotate":[p]}],invert:[{invert:[c]}],saturate:[{saturate:[x]}],sepia:[{sepia:[w]}],"backdrop-filter":[{"backdrop-filter":["","none"]}],"backdrop-blur":[{"backdrop-blur":[n]}],"backdrop-brightness":[{"backdrop-brightness":[r]}],"backdrop-contrast":[{"backdrop-contrast":[l]}],"backdrop-grayscale":[{"backdrop-grayscale":[u]}],"backdrop-hue-rotate":[{"backdrop-hue-rotate":[p]}],"backdrop-invert":[{"backdrop-invert":[c]}],"backdrop-opacity":[{"backdrop-opacity":[v]}],"backdrop-saturate":[{"backdrop-saturate":[x]}],"backdrop-sepia":[{"backdrop-sepia":[w]}],"border-collapse":[{border:["collapse","separate"]}],"border-spacing":[{"border-spacing":[i]}],"border-spacing-x":[{"border-spacing-x":[i]}],"border-spacing-y":[{"border-spacing-y":[i]}],"table-layout":[{table:["auto","fixed"]}],caption:[{caption:["top","bottom"]}],transition:[{transition:["none","all","","colors","opacity","shadow","transform",oe]}],duration:[{duration:W()}],ease:[{ease:["linear","in","out","in-out",oe]}],delay:[{delay:W()}],animate:[{animate:["none","spin","ping","pulse","bounce",oe]}],transform:[{transform:["","gpu","none"]}],scale:[{scale:[P]}],"scale-x":[{"scale-x":[P]}],"scale-y":[{"scale-y":[P]}],rotate:[{rotate:[$s,oe]}],"translate-x":[{"translate-x":[C]}],"translate-y":[{"translate-y":[C]}],"skew-x":[{"skew-x":[S]}],"skew-y":[{"skew-y":[S]}],"transform-origin":[{origin:["center","top","top-right","right","bottom-right","bottom","bottom-left","left","top-left",oe]}],accent:[{accent:["auto",e]}],appearance:[{appearance:["none","auto"]}],cursor:[{cursor:["auto","default","pointer","wait","text","move","help","not-allowed","none","context-menu","progress","cell","crosshair","vertical-text","alias","copy","no-drop","grab","grabbing","all-scroll","col-resize","row-resize","n-resize","e-resize","s-resize","w-resize","ne-resize","nw-resize","se-resize","sw-resize","ew-resize","ns-resize","nesw-resize","nwse-resize","zoom-in","zoom-out",oe]}],"caret-color":[{caret:[e]}],"pointer-events":[{"pointer-events":["none","auto"]}],resize:[{resize:["none","y","x",""]}],"scroll-behavior":[{scroll:["auto","smooth"]}],"scroll-m":[{"scroll-m":j()}],"scroll-mx":[{"scroll-mx":j()}],"scroll-my":[{"scroll-my":j()}],"scroll-ms":[{"scroll-ms":j()}],"scroll-me":[{"scroll-me":j()}],"scroll-mt":[{"scroll-mt":j()}],"scroll-mr":[{"scroll-mr":j()}],"scroll-mb":[{"scroll-mb":j()}],"scroll-ml":[{"scroll-ml":j()}],"scroll-p":[{"scroll-p":j()}],"scroll-px":[{"scroll-px":j()}],"scroll-py":[{"scroll-py":j()}],"scroll-ps":[{"scroll-ps":j()}],"scroll-pe":[{"scroll-pe":j()}],"scroll-pt":[{"scroll-pt":j()}],"scroll-pr":[{"scroll-pr":j()}],"scroll-pb":[{"scroll-pb":j()}],"scroll-pl":[{"scroll-pl":j()}],"snap-align":[{snap:["start","end","center","align-none"]}],"snap-stop":[{snap:["normal","always"]}],"snap-type":[{snap:["none","x","y","both"]}],"snap-strictness":[{snap:["mandatory","proximity"]}],touch:[{touch:["auto","none","manipulation"]}],"touch-x":[{"touch-pan":["x","left","right"]}],"touch-y":[{"touch-pan":["y","up","down"]}],"touch-pz":["touch-pinch-zoom"],select:[{select:["none","text","all","auto"]}],"will-change":[{"will-change":["auto","scroll","contents","transform",oe]}],fill:[{fill:[e,"none"]}],"stroke-w":[{stroke:[yr,Vr,hm]}],stroke:[{stroke:[e,"none"]}],sr:["sr-only","not-sr-only"],"forced-color-adjust":[{"forced-color-adjust":["auto","none"]}]},conflictingClassGroups:{overflow:["overflow-x","overflow-y"],overscroll:["overscroll-x","overscroll-y"],inset:["inset-x","inset-y","start","end","top","right","bottom","left"],"inset-x":["right","left"],"inset-y":["top","bottom"],flex:["basis","grow","shrink"],gap:["gap-x","gap-y"],p:["px","py","ps","pe","pt","pr","pb","pl"],px:["pr","pl"],py:["pt","pb"],m:["mx","my","ms","me","mt","mr","mb","ml"],mx:["mr","ml"],my:["mt","mb"],size:["w","h"],"font-size":["leading"],"fvn-normal":["fvn-ordinal","fvn-slashed-zero","fvn-figure","fvn-spacing","fvn-fraction"],"fvn-ordinal":["fvn-normal"],"fvn-slashed-zero":["fvn-normal"],"fvn-figure":["fvn-normal"],"fvn-spacing":["fvn-normal"],"fvn-fraction":["fvn-normal"],"line-clamp":["display","overflow"],rounded:["rounded-s","rounded-e","rounded-t","rounded-r","rounded-b","rounded-l","rounded-ss","rounded-se","rounded-ee","rounded-es","rounded-tl","rounded-tr","rounded-br","rounded-bl"],"rounded-s":["rounded-ss","rounded-es"],"rounded-e":["rounded-se","rounded-ee"],"rounded-t":["rounded-tl","rounded-tr"],"rounded-r":["rounded-tr","rounded-br"],"rounded-b":["rounded-br","rounded-bl"],"rounded-l":["rounded-tl","rounded-bl"],"border-spacing":["border-spacing-x","border-spacing-y"],"border-w":["border-w-s","border-w-e","border-w-t","border-w-r","border-w-b","border-w-l"],"border-w-x":["border-w-r","border-w-l"],"border-w-y":["border-w-t","border-w-b"],"border-color":["border-color-s","border-color-e","border-color-t","border-color-r","border-color-b","border-color-l"],"border-color-x":["border-color-r","border-color-l"],"border-color-y":["border-color-t","border-color-b"],"scroll-m":["scroll-mx","scroll-my","scroll-ms","scroll-me","scroll-mt","scroll-mr","scroll-mb","scroll-ml"],"scroll-mx":["scroll-mr","scroll-ml"],"scroll-my":["scroll-mt","scroll-mb"],"scroll-p":["scroll-px","scroll-py","scroll-ps","scroll-pe","scroll-pt","scroll-pr","scroll-pb","scroll-pl"],"scroll-px":["scroll-pr","scroll-pl"],"scroll-py":["scroll-pt","scroll-pb"],touch:["touch-x","touch-y","touch-pz"],"touch-x":["touch"],"touch-y":["touch"],"touch-pz":["touch"]},conflictingClassGroupModifiers:{"font-size":["leading"]}}},NT=dT(jT);function Se(...e){return NT(ue(e))}const $T=ME,MT=RE,B4=k.forwardRef(({className:e,children:t,...n},r)=>b.jsxs(v4,{ref:r,className:Se("flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1",e),...n,children:[t,b.jsx(IE,{asChild:!0,children:b.jsx(C4,{className:"h-4 w-4 opacity-50"})})]}));B4.displayName=v4.displayName;const z4=k.forwardRef(({className:e,...t},n)=>b.jsx(w4,{ref:n,className:Se("flex cursor-default items-center justify-center py-1",e),...t,children:b.jsx(qE,{className:"h-4 w-4"})}));z4.displayName=w4.displayName;const H4=k.forwardRef(({className:e,...t},n)=>b.jsx(b4,{ref:n,className:Se("flex cursor-default items-center justify-center py-1",e),...t,children:b.jsx(C4,{className:"h-4 w-4"})}));H4.displayName=b4.displayName;const G4=k.forwardRef(({className:e,children:t,position:n="popper",...r},a)=>(k.useEffect(()=>{const o=()=>{document.querySelectorAll("[data-radix-popper-content-wrapper]").forEach(l=>{const u=l;u.style.setProperty("--popover","0 0% 100%"),u.style.setProperty("--popover-foreground","222.2 84% 4.9%"),u.style.setProperty("--border","214.3 31.8% 91.4%"),u.style.setProperty("--background","0 0% 100%"),u.style.setProperty("--foreground","222.2 84% 4.9%"),(document.documentElement.classList.contains("dark")||document.body.classList.contains("dark"))&&(u.style.setProperty("--popover","222.2 84% 4.9%"),u.style.setProperty("--popover-foreground","210 40% 98%"),u.style.setProperty("--border","217.2 32.6% 17.5%"),u.style.setProperty("--background","222.2 84% 4.9%"),u.style.setProperty("--foreground","210 40% 98%"))})};o();const i=setTimeout(o,100);return()=>clearTimeout(i)}),b.jsx(DE,{children:b.jsxs(y4,{ref:a,className:Se("relative z-50 max-h-96 min-w-[8rem] overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-md data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",n==="popper"&&"data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1",e),position:n,...r,children:[b.jsx(z4,{}),b.jsx(LE,{className:Se("p-1",n==="popper"&&"h-[var(--radix-select-trigger-height)] w-full min-w-[var(--radix-select-trigger-width)]"),children:t}),b.jsx(H4,{})]})})));G4.displayName=y4.displayName;const RT=k.forwardRef(({className:e,...t},n)=>b.jsx(g4,{ref:n,className:Se("py-1.5 pl-8 pr-2 text-sm font-semibold",e),...t}));RT.displayName=g4.displayName;const U4=k.forwardRef(({className:e,children:t,...n},r)=>b.jsxs(x4,{ref:r,className:Se("relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",e),...n,children:[b.jsx("span",{className:"absolute left-2 flex h-3.5 w-3.5 items-center justify-center",children:b.jsx(BE,{children:b.jsx(WE,{className:"h-4 w-4"})})}),b.jsx(FE,{children:t})]}));U4.displayName=x4.displayName;const IT=k.forwardRef(({className:e,...t},n)=>b.jsx(P4,{ref:n,className:Se("-mx-1 my-1 h-px bg-muted",e),...t}));IT.displayName=P4.displayName;const jx=e=>typeof e=="boolean"?`${e}`:e===0?"0":e,Nx=ue,W4=(e,t)=>n=>{var r;if((t==null?void 0:t.variants)==null)return Nx(e,n==null?void 0:n.class,n==null?void 0:n.className);const{variants:a,defaultVariants:o}=t,i=Object.keys(a).map(u=>{const p=n==null?void 0:n[u],c=o==null?void 0:o[u];if(p===null)return null;const f=jx(p)||jx(c);return a[u][f]}),s=n&&Object.entries(n).reduce((u,p)=>{let[c,f]=p;return f===void 0||(u[c]=f),u},{}),l=t==null||(r=t.compoundVariants)===null||r===void 0?void 0:r.reduce((u,p)=>{let{class:c,className:f,...m}=p;return Object.entries(m).every(h=>{let[d,g]=h;return Array.isArray(g)?g.includes({...o,...s}[d]):{...o,...s}[d]===g})?[...u,c,f]:u},[]);return Nx(e,i,l,n==null?void 0:n.class,n==null?void 0:n.className)},DT=W4("inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",{variants:{variant:{default:"border-transparent bg-primary text-primary-foreground hover:bg-primary/80",secondary:"border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",destructive:"border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80",outline:"text-foreground"}},defaultVariants:{variant:"default"}});function LT({className:e,variant:t,...n}){return b.jsx("div",{className:Se(DT({variant:t}),e),...n})}const FT=W4("inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",{variants:{variant:{default:"bg-primary text-primary-foreground hover:bg-primary/90",destructive:"bg-destructive text-destructive-foreground hover:bg-destructive/90",outline:"border border-input bg-background hover:bg-accent hover:text-accent-foreground",secondary:"bg-secondary text-secondary-foreground hover:bg-secondary/80",ghost:"hover:bg-accent hover:text-accent-foreground",link:"text-primary underline-offset-4 hover:underline"},size:{default:"h-10 px-4 py-2",sm:"h-9 rounded-md px-3",lg:"h-11 rounded-md px-8",icon:"h-10 w-10"}},defaultVariants:{variant:"default",size:"default"}}),dp=k.forwardRef(({className:e,variant:t,size:n,asChild:r=!1,...a},o)=>{const i=r?LC:"button";return b.jsx(i,{className:Se(FT({variant:t,size:n,className:e})),ref:o,...a})});dp.displayName="Button";const BT=({selectedVersion:e,availableVersions:t,onVersionChange:n,metadata:r,onDownloadReport:a})=>{const o=p=>{try{return new Date(p).toLocaleDateString()}catch{return p}},i=p=>`https://github.com/envoyproxy/gateway/releases/tag/v${p}`,s=p=>`https://github.com/envoyproxy/gateway/releases/download/v${p}/benchmark_report.zip`,l=()=>{const p=i(e);window.open(p,"_blank","noopener,noreferrer")},u=()=>{const p=s(e);window.open(p,"_blank","noopener,noreferrer")};return b.jsxs("div",{className:"flex flex-col sm:flex-row items-start sm:items-center gap-3",children:[b.jsxs("div",{className:"flex items-center gap-2",children:[b.jsx("label",{className:"text-sm font-medium text-gray-700",children:"Version:"}),b.jsxs($T,{value:e,onValueChange:n,children:[b.jsx(B4,{className:"w-32",children:b.jsx(MT,{})}),b.jsx(G4,{children:t.map(p=>b.jsxs(U4,{value:p,children:["v",p]},p))})]})]}),r&&b.jsxs("div",{className:"flex flex-wrap items-center gap-2 text-xs text-gray-600",children:[b.jsxs("div",{className:"flex items-center gap-1",children:[b.jsx(UE,{className:"h-3 w-3"}),b.jsx("span",{children:o(r.date)})]}),r.environment&&b.jsx(LT,{variant:"secondary",className:"text-xs",children:r.environment}),r.gitCommit&&b.jsxs("div",{className:"flex items-center gap-1",children:[b.jsx(T4,{className:"h-3 w-3"}),b.jsx("span",{className:"font-mono",children:r.gitCommit.substring(0,7)})]}),r.description&&b.jsx("span",{className:"hidden sm:inline max-w-xs truncate",title:r.description,children:r.description})]}),b.jsxs("div",{className:"flex items-center gap-2 ml-auto",children:[b.jsxs(dp,{variant:"outline",size:"sm",onClick:l,className:"flex items-center gap-1 text-xs",children:[b.jsx(KE,{className:"h-3 w-3"}),"View Release"]}),b.jsxs(dp,{variant:"outline",size:"sm",onClick:u,className:"flex items-center gap-1 text-xs",children:[b.jsx(k4,{className:"h-3 w-3"}),"Download Benchmark"]})]})]})},ge=k.forwardRef(({className:e,...t},n)=>b.jsx("div",{ref:n,className:Se("rounded-lg border bg-card text-card-foreground shadow-sm",e),...t}));ge.displayName="Card";const Pe=k.forwardRef(({className:e,...t},n)=>b.jsx("div",{ref:n,className:Se("flex flex-col space-y-1.5 p-6",e),...t}));Pe.displayName="CardHeader";const xe=k.forwardRef(({className:e,...t},n)=>b.jsx("h3",{ref:n,className:Se("text-2xl font-semibold leading-none tracking-tight",e),...t}));xe.displayName="CardTitle";const en=k.forwardRef(({className:e,...t},n)=>b.jsx("p",{ref:n,className:Se("text-sm text-muted-foreground",e),...t}));en.displayName="CardDescription";const we=k.forwardRef(({className:e,...t},n)=>b.jsx("div",{ref:n,className:Se("p-6 pt-0",e),...t}));we.displayName="CardContent";const zT=k.forwardRef(({className:e,...t},n)=>b.jsx("div",{ref:n,className:Se("flex items-center p-6 pt-0",e),...t}));zT.displayName="CardFooter";const HT=({performanceSummary:e,benchmarkResults:t})=>{const n=Math.round(Math.max(...t.map(i=>i.throughput))),r=Math.round(e.avgLatency/1e3),a=Math.round(Math.max(...t.map(i=>i.resources.envoyGateway.memory.mean+i.resources.envoyProxy.memory.mean))),o=e.maxRoutes;return b.jsxs("div",{className:"grid grid-cols-1 md:grid-cols-4 gap-4 mb-8",children:[b.jsxs(ge,{className:"bg-gradient-to-r from-purple-600 to-indigo-600 text-white",children:[b.jsx(Pe,{className:"pb-2",children:b.jsxs(xe,{className:"text-sm font-medium flex items-center",children:[b.jsx(O4,{className:"h-4 w-4 mr-2"}),"Max Throughput in Test"]})}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[n.toLocaleString()," RPS"]}),b.jsx("p",{className:"text-purple-100 text-sm",children:"Requests per second"})]})]}),b.jsxs(ge,{className:"bg-gradient-to-r from-purple-600 to-indigo-600 text-white",children:[b.jsx(Pe,{className:"pb-2",children:b.jsxs(xe,{className:"text-sm font-medium flex items-center",children:[b.jsx(A4,{className:"h-4 w-4 mr-2"}),"Mean Response Time"]})}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[r,"ms"]}),b.jsx("p",{className:"text-purple-100 text-sm",children:"End-to-end via Nighthawk"})]})]}),b.jsxs(ge,{className:"bg-gradient-to-r from-purple-600 to-indigo-600 text-white",children:[b.jsx(Pe,{className:"pb-2",children:b.jsxs(xe,{className:"text-sm font-medium flex items-center",children:[b.jsx(YE,{className:"h-4 w-4 mr-2"}),"Memory Usage"]})}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[a,"MB"]}),b.jsxs("p",{className:"text-purple-100 text-sm",children:["Peak at ",o," routes"]})]})]}),b.jsxs(ge,{className:"bg-gradient-to-r from-purple-600 to-indigo-600 text-white",children:[b.jsx(Pe,{className:"pb-2",children:b.jsxs(xe,{className:"text-sm font-medium flex items-center text",children:[b.jsx(N4,{className:"h-4 w-4 mr-2"}),"Max Routes in Test"]})}),b.jsxs(we,{children:[b.jsx("div",{className:"text-2xl font-bold",children:o.toLocaleString()}),b.jsx("p",{className:"text-purple-100 text-sm",children:"HTTPRoutes tested"})]})]})]})};var GT=Array.isArray,qt=GT,UT=typeof Mu=="object"&&Mu&&Mu.Object===Object&&Mu,q4=UT,WT=q4,qT=typeof self=="object"&&self&&self.Object===Object&&self,VT=WT||qT||Function("return this")(),vr=VT,KT=vr,XT=KT.Symbol,Ou=XT,$x=Ou,V4=Object.prototype,YT=V4.hasOwnProperty,QT=V4.toString,Rs=$x?$x.toStringTag:void 0;function ZT(e){var t=YT.call(e,Rs),n=e[Rs];try{e[Rs]=void 0;var r=!0}catch{}var a=QT.call(e);return r&&(t?e[Rs]=n:delete e[Rs]),a}var JT=ZT,ej=Object.prototype,tj=ej.toString;function nj(e){return tj.call(e)}var rj=nj,Mx=Ou,aj=JT,oj=rj,ij="[object Null]",sj="[object Undefined]",Rx=Mx?Mx.toStringTag:void 0;function lj(e){return e==null?e===void 0?sj:ij:Rx&&Rx in Object(e)?aj(e):oj(e)}var Gr=lj;function uj(e){return e!=null&&typeof e=="object"}var Ur=uj,cj=Gr,pj=Ur,fj="[object Symbol]";function dj(e){return typeof e=="symbol"||pj(e)&&cj(e)==fj}var ps=dj,mj=qt,hj=ps,vj=/\.|\[(?:[^[\]]*|(["'])(?:(?!\1)[^\\]|\\.)*?\1)\]/,yj=/^\w*$/;function gj(e,t){if(mj(e))return!1;var n=typeof e;return n=="number"||n=="symbol"||n=="boolean"||e==null||hj(e)?!0:yj.test(e)||!vj.test(e)||t!=null&&e in Object(t)}var Ny=gj;function xj(e){var t=typeof e;return e!=null&&(t=="object"||t=="function")}var Na=xj;const fs=_e(Na);var wj=Gr,bj=Na,Pj="[object AsyncFunction]",Sj="[object Function]",Oj="[object GeneratorFunction]",kj="[object Proxy]";function Cj(e){if(!bj(e))return!1;var t=wj(e);return t==Sj||t==Oj||t==Pj||t==kj}var $y=Cj;const ae=_e($y);var _j=vr,Aj=_j["__core-js_shared__"],Ej=Aj,vm=Ej,Ix=function(){var e=/[^.]+$/.exec(vm&&vm.keys&&vm.keys.IE_PROTO||"");return e?"Symbol(src)_1."+e:""}();function Tj(e){return!!Ix&&Ix in e}var jj=Tj,Nj=Function.prototype,$j=Nj.toString;function Mj(e){if(e!=null){try{return $j.call(e)}catch{}try{return e+""}catch{}}return""}var K4=Mj,Rj=$y,Ij=jj,Dj=Na,Lj=K4,Fj=/[\\^$.*+?()[\]{}|]/g,Bj=/^\[object .+?Constructor\]$/,zj=Function.prototype,Hj=Object.prototype,Gj=zj.toString,Uj=Hj.hasOwnProperty,Wj=RegExp("^"+Gj.call(Uj).replace(Fj,"\\$&").replace(/hasOwnProperty|(function).*?(?=\\\()| for .+?(?=\\\])/g,"$1.*?")+"$");function qj(e){if(!Dj(e)||Ij(e))return!1;var t=Rj(e)?Wj:Bj;return t.test(Lj(e))}var Vj=qj;function Kj(e,t){return e==null?void 0:e[t]}var Xj=Kj,Yj=Vj,Qj=Xj;function Zj(e,t){var n=Qj(e,t);return Yj(n)?n:void 0}var Co=Zj,Jj=Co,eN=Jj(Object,"create"),zf=eN,Dx=zf;function tN(){this.__data__=Dx?Dx(null):{},this.size=0}var nN=tN;function rN(e){var t=this.has(e)&&delete this.__data__[e];return this.size-=t?1:0,t}var aN=rN,oN=zf,iN="__lodash_hash_undefined__",sN=Object.prototype,lN=sN.hasOwnProperty;function uN(e){var t=this.__data__;if(oN){var n=t[e];return n===iN?void 0:n}return lN.call(t,e)?t[e]:void 0}var cN=uN,pN=zf,fN=Object.prototype,dN=fN.hasOwnProperty;function mN(e){var t=this.__data__;return pN?t[e]!==void 0:dN.call(t,e)}var hN=mN,vN=zf,yN="__lodash_hash_undefined__";function gN(e,t){var n=this.__data__;return this.size+=this.has(e)?0:1,n[e]=vN&&t===void 0?yN:t,this}var xN=gN,wN=nN,bN=aN,PN=cN,SN=hN,ON=xN;function ds(e){var t=-1,n=e==null?0:e.length;for(this.clear();++t<n;){var r=e[t];this.set(r[0],r[1])}}ds.prototype.clear=wN;ds.prototype.delete=bN;ds.prototype.get=PN;ds.prototype.has=SN;ds.prototype.set=ON;var kN=ds;function CN(){this.__data__=[],this.size=0}var _N=CN;function AN(e,t){return e===t||e!==e&&t!==t}var My=AN,EN=My;function TN(e,t){for(var n=e.length;n--;)if(EN(e[n][0],t))return n;return-1}var Hf=TN,jN=Hf,NN=Array.prototype,$N=NN.splice;function MN(e){var t=this.__data__,n=jN(t,e);if(n<0)return!1;var r=t.length-1;return n==r?t.pop():$N.call(t,n,1),--this.size,!0}var RN=MN,IN=Hf;function DN(e){var t=this.__data__,n=IN(t,e);return n<0?void 0:t[n][1]}var LN=DN,FN=Hf;function BN(e){return FN(this.__data__,e)>-1}var zN=BN,HN=Hf;function GN(e,t){var n=this.__data__,r=HN(n,e);return r<0?(++this.size,n.push([e,t])):n[r][1]=t,this}var UN=GN,WN=_N,qN=RN,VN=LN,KN=zN,XN=UN;function ms(e){var t=-1,n=e==null?0:e.length;for(this.clear();++t<n;){var r=e[t];this.set(r[0],r[1])}}ms.prototype.clear=WN;ms.prototype.delete=qN;ms.prototype.get=VN;ms.prototype.has=KN;ms.prototype.set=XN;var Gf=ms,YN=Co,QN=vr,ZN=YN(QN,"Map"),Ry=ZN,Lx=kN,JN=Gf,e$=Ry;function t$(){this.size=0,this.__data__={hash:new Lx,map:new(e$||JN),string:new Lx}}var n$=t$;function r$(e){var t=typeof e;return t=="string"||t=="number"||t=="symbol"||t=="boolean"?e!=="__proto__":e===null}var a$=r$,o$=a$;function i$(e,t){var n=e.__data__;return o$(t)?n[typeof t=="string"?"string":"hash"]:n.map}var Uf=i$,s$=Uf;function l$(e){var t=s$(this,e).delete(e);return this.size-=t?1:0,t}var u$=l$,c$=Uf;function p$(e){return c$(this,e).get(e)}var f$=p$,d$=Uf;function m$(e){return d$(this,e).has(e)}var h$=m$,v$=Uf;function y$(e,t){var n=v$(this,e),r=n.size;return n.set(e,t),this.size+=n.size==r?0:1,this}var g$=y$,x$=n$,w$=u$,b$=f$,P$=h$,S$=g$;function hs(e){var t=-1,n=e==null?0:e.length;for(this.clear();++t<n;){var r=e[t];this.set(r[0],r[1])}}hs.prototype.clear=x$;hs.prototype.delete=w$;hs.prototype.get=b$;hs.prototype.has=P$;hs.prototype.set=S$;var Iy=hs,X4=Iy,O$="Expected a function";function Dy(e,t){if(typeof e!="function"||t!=null&&typeof t!="function")throw new TypeError(O$);var n=function(){var r=arguments,a=t?t.apply(this,r):r[0],o=n.cache;if(o.has(a))return o.get(a);var i=e.apply(this,r);return n.cache=o.set(a,i)||o,i};return n.cache=new(Dy.Cache||X4),n}Dy.Cache=X4;var Y4=Dy;const k$=_e(Y4);var C$=Y4,_$=500;function A$(e){var t=C$(e,function(r){return n.size===_$&&n.clear(),r}),n=t.cache;return t}var E$=A$,T$=E$,j$=/[^.[\]]+|\[(?:(-?\d+(?:\.\d+)?)|(["'])((?:(?!\2)[^\\]|\\.)*?)\2)\]|(?=(?:\.|\[\])(?:\.|\[\]|$))/g,N$=/\\(\\)?/g,$$=T$(function(e){var t=[];return e.charCodeAt(0)===46&&t.push(""),e.replace(j$,function(n,r,a,o){t.push(a?o.replace(N$,"$1"):r||n)}),t}),M$=$$;function R$(e,t){for(var n=-1,r=e==null?0:e.length,a=Array(r);++n<r;)a[n]=t(e[n],n,e);return a}var Ly=R$,Fx=Ou,I$=Ly,D$=qt,L$=ps,Bx=Fx?Fx.prototype:void 0,zx=Bx?Bx.toString:void 0;function Q4(e){if(typeof e=="string")return e;if(D$(e))return I$(e,Q4)+"";if(L$(e))return zx?zx.call(e):"";var t=e+"";return t=="0"&&1/e==-1/0?"-0":t}var F$=Q4,B$=F$;function z$(e){return e==null?"":B$(e)}var Z4=z$,H$=qt,G$=Ny,U$=M$,W$=Z4;function q$(e,t){return H$(e)?e:G$(e,t)?[e]:U$(W$(e))}var J4=q$,V$=ps;function K$(e){if(typeof e=="string"||V$(e))return e;var t=e+"";return t=="0"&&1/e==-1/0?"-0":t}var Wf=K$,X$=J4,Y$=Wf;function Q$(e,t){t=X$(t,e);for(var n=0,r=t.length;e!=null&&n<r;)e=e[Y$(t[n++])];return n&&n==r?e:void 0}var Fy=Q$,Z$=Fy;function J$(e,t,n){var r=e==null?void 0:Z$(e,t);return r===void 0?n:r}var e8=J$;const vn=_e(e8);function eM(e){return e==null}var tM=eM;const le=_e(tM);var nM=Gr,rM=qt,aM=Ur,oM="[object String]";function iM(e){return typeof e=="string"||!rM(e)&&aM(e)&&nM(e)==oM}var sM=iM;const xo=_e(sM);var t8={exports:{}},ke={};/**
 * @license React
 * react-is.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var By=Symbol.for("react.element"),zy=Symbol.for("react.portal"),qf=Symbol.for("react.fragment"),Vf=Symbol.for("react.strict_mode"),Kf=Symbol.for("react.profiler"),Xf=Symbol.for("react.provider"),Yf=Symbol.for("react.context"),lM=Symbol.for("react.server_context"),Qf=Symbol.for("react.forward_ref"),Zf=Symbol.for("react.suspense"),Jf=Symbol.for("react.suspense_list"),ed=Symbol.for("react.memo"),td=Symbol.for("react.lazy"),uM=Symbol.for("react.offscreen"),n8;n8=Symbol.for("react.module.reference");function bn(e){if(typeof e=="object"&&e!==null){var t=e.$$typeof;switch(t){case By:switch(e=e.type,e){case qf:case Kf:case Vf:case Zf:case Jf:return e;default:switch(e=e&&e.$$typeof,e){case lM:case Yf:case Qf:case td:case ed:case Xf:return e;default:return t}}case zy:return t}}}ke.ContextConsumer=Yf;ke.ContextProvider=Xf;ke.Element=By;ke.ForwardRef=Qf;ke.Fragment=qf;ke.Lazy=td;ke.Memo=ed;ke.Portal=zy;ke.Profiler=Kf;ke.StrictMode=Vf;ke.Suspense=Zf;ke.SuspenseList=Jf;ke.isAsyncMode=function(){return!1};ke.isConcurrentMode=function(){return!1};ke.isContextConsumer=function(e){return bn(e)===Yf};ke.isContextProvider=function(e){return bn(e)===Xf};ke.isElement=function(e){return typeof e=="object"&&e!==null&&e.$$typeof===By};ke.isForwardRef=function(e){return bn(e)===Qf};ke.isFragment=function(e){return bn(e)===qf};ke.isLazy=function(e){return bn(e)===td};ke.isMemo=function(e){return bn(e)===ed};ke.isPortal=function(e){return bn(e)===zy};ke.isProfiler=function(e){return bn(e)===Kf};ke.isStrictMode=function(e){return bn(e)===Vf};ke.isSuspense=function(e){return bn(e)===Zf};ke.isSuspenseList=function(e){return bn(e)===Jf};ke.isValidElementType=function(e){return typeof e=="string"||typeof e=="function"||e===qf||e===Kf||e===Vf||e===Zf||e===Jf||e===uM||typeof e=="object"&&e!==null&&(e.$$typeof===td||e.$$typeof===ed||e.$$typeof===Xf||e.$$typeof===Yf||e.$$typeof===Qf||e.$$typeof===n8||e.getModuleId!==void 0)};ke.typeOf=bn;t8.exports=ke;var cM=t8.exports,pM=Gr,fM=Ur,dM="[object Number]";function mM(e){return typeof e=="number"||fM(e)&&pM(e)==dM}var r8=mM;const hM=_e(r8);var vM=r8;function yM(e){return vM(e)&&e!=+e}var gM=yM;const vs=_e(gM);var In=function(t){return t===0?0:t>0?1:-1},Xa=function(t){return xo(t)&&t.indexOf("%")===t.length-1},V=function(t){return hM(t)&&!vs(t)},xM=function(t){return le(t)},ot=function(t){return V(t)||xo(t)},wM=0,ys=function(t){var n=++wM;return"".concat(t||"").concat(n)},wo=function(t,n){var r=arguments.length>2&&arguments[2]!==void 0?arguments[2]:0,a=arguments.length>3&&arguments[3]!==void 0?arguments[3]:!1;if(!V(t)&&!xo(t))return r;var o;if(Xa(t)){var i=t.indexOf("%");o=n*parseFloat(t.slice(0,i))/100}else o=+t;return vs(o)&&(o=r),a&&o>n&&(o=n),o},ra=function(t){if(!t)return null;var n=Object.keys(t);return n&&n.length?t[n[0]]:null},bM=function(t){if(!Array.isArray(t))return!1;for(var n=t.length,r={},a=0;a<n;a++)if(!r[t[a]])r[t[a]]=!0;else return!0;return!1},dt=function(t,n){return V(t)&&V(n)?function(r){return t+r*(n-t)}:function(){return n}};function mp(e,t,n){return!e||!e.length?null:e.find(function(r){return r&&(typeof t=="function"?t(r):vn(r,t))===n})}var PM=function(t,n){return V(t)&&V(n)?t-n:xo(t)&&xo(n)?t.localeCompare(n):t instanceof Date&&n instanceof Date?t.getTime()-n.getTime():String(t).localeCompare(String(n))};function ui(e,t){for(var n in e)if({}.hasOwnProperty.call(e,n)&&(!{}.hasOwnProperty.call(t,n)||e[n]!==t[n]))return!1;for(var r in t)if({}.hasOwnProperty.call(t,r)&&!{}.hasOwnProperty.call(e,r))return!1;return!0}function Xh(e){"@babel/helpers - typeof";return Xh=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Xh(e)}var SM=["viewBox","children"],OM=["aria-activedescendant","aria-atomic","aria-autocomplete","aria-busy","aria-checked","aria-colcount","aria-colindex","aria-colspan","aria-controls","aria-current","aria-describedby","aria-details","aria-disabled","aria-errormessage","aria-expanded","aria-flowto","aria-haspopup","aria-hidden","aria-invalid","aria-keyshortcuts","aria-label","aria-labelledby","aria-level","aria-live","aria-modal","aria-multiline","aria-multiselectable","aria-orientation","aria-owns","aria-placeholder","aria-posinset","aria-pressed","aria-readonly","aria-relevant","aria-required","aria-roledescription","aria-rowcount","aria-rowindex","aria-rowspan","aria-selected","aria-setsize","aria-sort","aria-valuemax","aria-valuemin","aria-valuenow","aria-valuetext","className","color","height","id","lang","max","media","method","min","name","style","target","width","role","tabIndex","accentHeight","accumulate","additive","alignmentBaseline","allowReorder","alphabetic","amplitude","arabicForm","ascent","attributeName","attributeType","autoReverse","azimuth","baseFrequency","baselineShift","baseProfile","bbox","begin","bias","by","calcMode","capHeight","clip","clipPath","clipPathUnits","clipRule","colorInterpolation","colorInterpolationFilters","colorProfile","colorRendering","contentScriptType","contentStyleType","cursor","cx","cy","d","decelerate","descent","diffuseConstant","direction","display","divisor","dominantBaseline","dur","dx","dy","edgeMode","elevation","enableBackground","end","exponent","externalResourcesRequired","fill","fillOpacity","fillRule","filter","filterRes","filterUnits","floodColor","floodOpacity","focusable","fontFamily","fontSize","fontSizeAdjust","fontStretch","fontStyle","fontVariant","fontWeight","format","from","fx","fy","g1","g2","glyphName","glyphOrientationHorizontal","glyphOrientationVertical","glyphRef","gradientTransform","gradientUnits","hanging","horizAdvX","horizOriginX","href","ideographic","imageRendering","in2","in","intercept","k1","k2","k3","k4","k","kernelMatrix","kernelUnitLength","kerning","keyPoints","keySplines","keyTimes","lengthAdjust","letterSpacing","lightingColor","limitingConeAngle","local","markerEnd","markerHeight","markerMid","markerStart","markerUnits","markerWidth","mask","maskContentUnits","maskUnits","mathematical","mode","numOctaves","offset","opacity","operator","order","orient","orientation","origin","overflow","overlinePosition","overlineThickness","paintOrder","panose1","pathLength","patternContentUnits","patternTransform","patternUnits","pointerEvents","pointsAtX","pointsAtY","pointsAtZ","preserveAlpha","preserveAspectRatio","primitiveUnits","r","radius","refX","refY","renderingIntent","repeatCount","repeatDur","requiredExtensions","requiredFeatures","restart","result","rotate","rx","ry","seed","shapeRendering","slope","spacing","specularConstant","specularExponent","speed","spreadMethod","startOffset","stdDeviation","stemh","stemv","stitchTiles","stopColor","stopOpacity","strikethroughPosition","strikethroughThickness","string","stroke","strokeDasharray","strokeDashoffset","strokeLinecap","strokeLinejoin","strokeMiterlimit","strokeOpacity","strokeWidth","surfaceScale","systemLanguage","tableValues","targetX","targetY","textAnchor","textDecoration","textLength","textRendering","to","transform","u1","u2","underlinePosition","underlineThickness","unicode","unicodeBidi","unicodeRange","unitsPerEm","vAlphabetic","values","vectorEffect","version","vertAdvY","vertOriginX","vertOriginY","vHanging","vIdeographic","viewTarget","visibility","vMathematical","widths","wordSpacing","writingMode","x1","x2","x","xChannelSelector","xHeight","xlinkActuate","xlinkArcrole","xlinkHref","xlinkRole","xlinkShow","xlinkTitle","xlinkType","xmlBase","xmlLang","xmlns","xmlnsXlink","xmlSpace","y1","y2","y","yChannelSelector","z","zoomAndPan","ref","key","angle"],Hx=["points","pathLength"],ym={svg:SM,polygon:Hx,polyline:Hx},Hy=["dangerouslySetInnerHTML","onCopy","onCopyCapture","onCut","onCutCapture","onPaste","onPasteCapture","onCompositionEnd","onCompositionEndCapture","onCompositionStart","onCompositionStartCapture","onCompositionUpdate","onCompositionUpdateCapture","onFocus","onFocusCapture","onBlur","onBlurCapture","onChange","onChangeCapture","onBeforeInput","onBeforeInputCapture","onInput","onInputCapture","onReset","onResetCapture","onSubmit","onSubmitCapture","onInvalid","onInvalidCapture","onLoad","onLoadCapture","onError","onErrorCapture","onKeyDown","onKeyDownCapture","onKeyPress","onKeyPressCapture","onKeyUp","onKeyUpCapture","onAbort","onAbortCapture","onCanPlay","onCanPlayCapture","onCanPlayThrough","onCanPlayThroughCapture","onDurationChange","onDurationChangeCapture","onEmptied","onEmptiedCapture","onEncrypted","onEncryptedCapture","onEnded","onEndedCapture","onLoadedData","onLoadedDataCapture","onLoadedMetadata","onLoadedMetadataCapture","onLoadStart","onLoadStartCapture","onPause","onPauseCapture","onPlay","onPlayCapture","onPlaying","onPlayingCapture","onProgress","onProgressCapture","onRateChange","onRateChangeCapture","onSeeked","onSeekedCapture","onSeeking","onSeekingCapture","onStalled","onStalledCapture","onSuspend","onSuspendCapture","onTimeUpdate","onTimeUpdateCapture","onVolumeChange","onVolumeChangeCapture","onWaiting","onWaitingCapture","onAuxClick","onAuxClickCapture","onClick","onClickCapture","onContextMenu","onContextMenuCapture","onDoubleClick","onDoubleClickCapture","onDrag","onDragCapture","onDragEnd","onDragEndCapture","onDragEnter","onDragEnterCapture","onDragExit","onDragExitCapture","onDragLeave","onDragLeaveCapture","onDragOver","onDragOverCapture","onDragStart","onDragStartCapture","onDrop","onDropCapture","onMouseDown","onMouseDownCapture","onMouseEnter","onMouseLeave","onMouseMove","onMouseMoveCapture","onMouseOut","onMouseOutCapture","onMouseOver","onMouseOverCapture","onMouseUp","onMouseUpCapture","onSelect","onSelectCapture","onTouchCancel","onTouchCancelCapture","onTouchEnd","onTouchEndCapture","onTouchMove","onTouchMoveCapture","onTouchStart","onTouchStartCapture","onPointerDown","onPointerDownCapture","onPointerMove","onPointerMoveCapture","onPointerUp","onPointerUpCapture","onPointerCancel","onPointerCancelCapture","onPointerEnter","onPointerEnterCapture","onPointerLeave","onPointerLeaveCapture","onPointerOver","onPointerOverCapture","onPointerOut","onPointerOutCapture","onGotPointerCapture","onGotPointerCaptureCapture","onLostPointerCapture","onLostPointerCaptureCapture","onScroll","onScrollCapture","onWheel","onWheelCapture","onAnimationStart","onAnimationStartCapture","onAnimationEnd","onAnimationEndCapture","onAnimationIteration","onAnimationIterationCapture","onTransitionEnd","onTransitionEndCapture"],hp=function(t,n){if(!t||typeof t=="function"||typeof t=="boolean")return null;var r=t;if(k.isValidElement(t)&&(r=t.props),!fs(r))return null;var a={};return Object.keys(r).forEach(function(o){Hy.includes(o)&&(a[o]=n||function(i){return r[o](r,i)})}),a},kM=function(t,n,r){return function(a){return t(n,r,a),null}},vp=function(t,n,r){if(!fs(t)||Xh(t)!=="object")return null;var a=null;return Object.keys(t).forEach(function(o){var i=t[o];Hy.includes(o)&&typeof i=="function"&&(a||(a={}),a[o]=kM(i,n,r))}),a},CM=["children"],_M=["children"];function Gx(e,t){if(e==null)return{};var n=AM(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function AM(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function Yh(e){"@babel/helpers - typeof";return Yh=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Yh(e)}var Ux={click:"onClick",mousedown:"onMouseDown",mouseup:"onMouseUp",mouseover:"onMouseOver",mousemove:"onMouseMove",mouseout:"onMouseOut",mouseenter:"onMouseEnter",mouseleave:"onMouseLeave",touchcancel:"onTouchCancel",touchend:"onTouchEnd",touchmove:"onTouchMove",touchstart:"onTouchStart",contextmenu:"onContextMenu",dblclick:"onDoubleClick"},Er=function(t){return typeof t=="string"?t:t?t.displayName||t.name||"Component":""},Wx=null,gm=null,Gy=function e(t){if(t===Wx&&Array.isArray(gm))return gm;var n=[];return k.Children.forEach(t,function(r){le(r)||(cM.isFragment(r)?n=n.concat(e(r.props.children)):n.push(r))}),gm=n,Wx=t,n};function yn(e,t){var n=[],r=[];return Array.isArray(t)?r=t.map(function(a){return Er(a)}):r=[Er(t)],Gy(e).forEach(function(a){var o=vn(a,"type.displayName")||vn(a,"type.name");r.indexOf(o)!==-1&&n.push(a)}),n}function Yt(e,t){var n=yn(e,t);return n&&n[0]}var qx=function(t){if(!t||!t.props)return!1;var n=t.props,r=n.width,a=n.height;return!(!V(r)||r<=0||!V(a)||a<=0)},EM=["a","altGlyph","altGlyphDef","altGlyphItem","animate","animateColor","animateMotion","animateTransform","circle","clipPath","color-profile","cursor","defs","desc","ellipse","feBlend","feColormatrix","feComponentTransfer","feComposite","feConvolveMatrix","feDiffuseLighting","feDisplacementMap","feDistantLight","feFlood","feFuncA","feFuncB","feFuncG","feFuncR","feGaussianBlur","feImage","feMerge","feMergeNode","feMorphology","feOffset","fePointLight","feSpecularLighting","feSpotLight","feTile","feTurbulence","filter","font","font-face","font-face-format","font-face-name","font-face-url","foreignObject","g","glyph","glyphRef","hkern","image","line","lineGradient","marker","mask","metadata","missing-glyph","mpath","path","pattern","polygon","polyline","radialGradient","rect","script","set","stop","style","svg","switch","symbol","text","textPath","title","tref","tspan","use","view","vkern"],TM=function(t){return t&&t.type&&xo(t.type)&&EM.indexOf(t.type)>=0},a8=function(t){return t&&Yh(t)==="object"&&"clipDot"in t},jM=function(t,n,r,a){var o,i=(o=ym==null?void 0:ym[a])!==null&&o!==void 0?o:[];return n.startsWith("data-")||!ae(t)&&(a&&i.includes(n)||OM.includes(n))||r&&Hy.includes(n)},ie=function(t,n,r){if(!t||typeof t=="function"||typeof t=="boolean")return null;var a=t;if(k.isValidElement(t)&&(a=t.props),!fs(a))return null;var o={};return Object.keys(a).forEach(function(i){var s;jM((s=a)===null||s===void 0?void 0:s[i],i,n,r)&&(o[i]=a[i])}),o},Qh=function e(t,n){if(t===n)return!0;var r=k.Children.count(t);if(r!==k.Children.count(n))return!1;if(r===0)return!0;if(r===1)return Vx(Array.isArray(t)?t[0]:t,Array.isArray(n)?n[0]:n);for(var a=0;a<r;a++){var o=t[a],i=n[a];if(Array.isArray(o)||Array.isArray(i)){if(!e(o,i))return!1}else if(!Vx(o,i))return!1}return!0},Vx=function(t,n){if(le(t)&&le(n))return!0;if(!le(t)&&!le(n)){var r=t.props||{},a=r.children,o=Gx(r,CM),i=n.props||{},s=i.children,l=Gx(i,_M);return a&&s?ui(o,l)&&Qh(a,s):!a&&!s?ui(o,l):!1}return!1},Kx=function(t,n){var r=[],a={};return Gy(t).forEach(function(o,i){if(TM(o))r.push(o);else if(o){var s=Er(o.type),l=n[s]||{},u=l.handler,p=l.once;if(u&&(!p||!a[s])){var c=u(o,s,i);r.push(c),a[s]=!0}}}),r},NM=function(t){var n=t&&t.type;return n&&Ux[n]?Ux[n]:null},$M=function(t,n){return Gy(n).indexOf(t)},MM=["children","width","height","viewBox","className","style","title","desc"];function Zh(){return Zh=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Zh.apply(this,arguments)}function RM(e,t){if(e==null)return{};var n=IM(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function IM(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function Jh(e){var t=e.children,n=e.width,r=e.height,a=e.viewBox,o=e.className,i=e.style,s=e.title,l=e.desc,u=RM(e,MM),p=a||{width:n,height:r,x:0,y:0},c=ue("recharts-surface",o);return E.createElement("svg",Zh({},ie(u,!0,"svg"),{className:c,width:n,height:r,style:i,viewBox:"".concat(p.x," ").concat(p.y," ").concat(p.width," ").concat(p.height)}),E.createElement("title",null,s),E.createElement("desc",null,l),t)}var DM=["children","className"];function e0(){return e0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},e0.apply(this,arguments)}function LM(e,t){if(e==null)return{};var n=FM(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function FM(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}var Ne=E.forwardRef(function(e,t){var n=e.children,r=e.className,a=LM(e,DM),o=ue("recharts-layer",r);return E.createElement("g",e0({className:o},ie(a,!0),{ref:t}),n)}),Tr=function(t,n){for(var r=arguments.length,a=new Array(r>2?r-2:0),o=2;o<r;o++)a[o-2]=arguments[o]};function BM(e,t,n){var r=-1,a=e.length;t<0&&(t=-t>a?0:a+t),n=n>a?a:n,n<0&&(n+=a),a=t>n?0:n-t>>>0,t>>>=0;for(var o=Array(a);++r<a;)o[r]=e[r+t];return o}var zM=BM,HM=zM;function GM(e,t,n){var r=e.length;return n=n===void 0?r:n,!t&&n>=r?e:HM(e,t,n)}var UM=GM,WM="\\ud800-\\udfff",qM="\\u0300-\\u036f",VM="\\ufe20-\\ufe2f",KM="\\u20d0-\\u20ff",XM=qM+VM+KM,YM="\\ufe0e\\ufe0f",QM="\\u200d",ZM=RegExp("["+QM+WM+XM+YM+"]");function JM(e){return ZM.test(e)}var o8=JM;function eR(e){return e.split("")}var tR=eR,i8="\\ud800-\\udfff",nR="\\u0300-\\u036f",rR="\\ufe20-\\ufe2f",aR="\\u20d0-\\u20ff",oR=nR+rR+aR,iR="\\ufe0e\\ufe0f",sR="["+i8+"]",t0="["+oR+"]",n0="\\ud83c[\\udffb-\\udfff]",lR="(?:"+t0+"|"+n0+")",s8="[^"+i8+"]",l8="(?:\\ud83c[\\udde6-\\uddff]){2}",u8="[\\ud800-\\udbff][\\udc00-\\udfff]",uR="\\u200d",c8=lR+"?",p8="["+iR+"]?",cR="(?:"+uR+"(?:"+[s8,l8,u8].join("|")+")"+p8+c8+")*",pR=p8+c8+cR,fR="(?:"+[s8+t0+"?",t0,l8,u8,sR].join("|")+")",dR=RegExp(n0+"(?="+n0+")|"+fR+pR,"g");function mR(e){return e.match(dR)||[]}var hR=mR,vR=tR,yR=o8,gR=hR;function xR(e){return yR(e)?gR(e):vR(e)}var wR=xR,bR=UM,PR=o8,SR=wR,OR=Z4;function kR(e){return function(t){t=OR(t);var n=PR(t)?SR(t):void 0,r=n?n[0]:t.charAt(0),a=n?bR(n,1).join(""):t.slice(1);return r[e]()+a}}var CR=kR,_R=CR,AR=_R("toUpperCase"),ER=AR;const nd=_e(ER);function Te(e){return function(){return e}}const f8=Math.cos,yp=Math.sin,qn=Math.sqrt,gp=Math.PI,rd=2*gp,r0=Math.PI,a0=2*r0,Ga=1e-6,TR=a0-Ga;function d8(e){this._+=e[0];for(let t=1,n=e.length;t<n;++t)this._+=arguments[t]+e[t]}function jR(e){let t=Math.floor(e);if(!(t>=0))throw new Error(`invalid digits: ${e}`);if(t>15)return d8;const n=10**t;return function(r){this._+=r[0];for(let a=1,o=r.length;a<o;++a)this._+=Math.round(arguments[a]*n)/n+r[a]}}class NR{constructor(t){this._x0=this._y0=this._x1=this._y1=null,this._="",this._append=t==null?d8:jR(t)}moveTo(t,n){this._append`M${this._x0=this._x1=+t},${this._y0=this._y1=+n}`}closePath(){this._x1!==null&&(this._x1=this._x0,this._y1=this._y0,this._append`Z`)}lineTo(t,n){this._append`L${this._x1=+t},${this._y1=+n}`}quadraticCurveTo(t,n,r,a){this._append`Q${+t},${+n},${this._x1=+r},${this._y1=+a}`}bezierCurveTo(t,n,r,a,o,i){this._append`C${+t},${+n},${+r},${+a},${this._x1=+o},${this._y1=+i}`}arcTo(t,n,r,a,o){if(t=+t,n=+n,r=+r,a=+a,o=+o,o<0)throw new Error(`negative radius: ${o}`);let i=this._x1,s=this._y1,l=r-t,u=a-n,p=i-t,c=s-n,f=p*p+c*c;if(this._x1===null)this._append`M${this._x1=t},${this._y1=n}`;else if(f>Ga)if(!(Math.abs(c*l-u*p)>Ga)||!o)this._append`L${this._x1=t},${this._y1=n}`;else{let m=r-i,h=a-s,d=l*l+u*u,g=m*m+h*h,v=Math.sqrt(d),y=Math.sqrt(f),x=o*Math.tan((r0-Math.acos((d+f-g)/(2*v*y)))/2),P=x/y,w=x/v;Math.abs(P-1)>Ga&&this._append`L${t+P*p},${n+P*c}`,this._append`A${o},${o},0,0,${+(c*m>p*h)},${this._x1=t+w*l},${this._y1=n+w*u}`}}arc(t,n,r,a,o,i){if(t=+t,n=+n,r=+r,i=!!i,r<0)throw new Error(`negative radius: ${r}`);let s=r*Math.cos(a),l=r*Math.sin(a),u=t+s,p=n+l,c=1^i,f=i?a-o:o-a;this._x1===null?this._append`M${u},${p}`:(Math.abs(this._x1-u)>Ga||Math.abs(this._y1-p)>Ga)&&this._append`L${u},${p}`,r&&(f<0&&(f=f%a0+a0),f>TR?this._append`A${r},${r},0,1,${c},${t-s},${n-l}A${r},${r},0,1,${c},${this._x1=u},${this._y1=p}`:f>Ga&&this._append`A${r},${r},0,${+(f>=r0)},${c},${this._x1=t+r*Math.cos(o)},${this._y1=n+r*Math.sin(o)}`)}rect(t,n,r,a){this._append`M${this._x0=this._x1=+t},${this._y0=this._y1=+n}h${r=+r}v${+a}h${-r}Z`}toString(){return this._}}function Uy(e){let t=3;return e.digits=function(n){if(!arguments.length)return t;if(n==null)t=null;else{const r=Math.floor(n);if(!(r>=0))throw new RangeError(`invalid digits: ${n}`);t=r}return e},()=>new NR(t)}function Wy(e){return typeof e=="object"&&"length"in e?e:Array.from(e)}function m8(e){this._context=e}m8.prototype={areaStart:function(){this._line=0},areaEnd:function(){this._line=NaN},lineStart:function(){this._point=0},lineEnd:function(){(this._line||this._line!==0&&this._point===1)&&this._context.closePath(),this._line=1-this._line},point:function(e,t){switch(e=+e,t=+t,this._point){case 0:this._point=1,this._line?this._context.lineTo(e,t):this._context.moveTo(e,t);break;case 1:this._point=2;default:this._context.lineTo(e,t);break}}};function ad(e){return new m8(e)}function h8(e){return e[0]}function v8(e){return e[1]}function y8(e,t){var n=Te(!0),r=null,a=ad,o=null,i=Uy(s);e=typeof e=="function"?e:e===void 0?h8:Te(e),t=typeof t=="function"?t:t===void 0?v8:Te(t);function s(l){var u,p=(l=Wy(l)).length,c,f=!1,m;for(r==null&&(o=a(m=i())),u=0;u<=p;++u)!(u<p&&n(c=l[u],u,l))===f&&((f=!f)?o.lineStart():o.lineEnd()),f&&o.point(+e(c,u,l),+t(c,u,l));if(m)return o=null,m+""||null}return s.x=function(l){return arguments.length?(e=typeof l=="function"?l:Te(+l),s):e},s.y=function(l){return arguments.length?(t=typeof l=="function"?l:Te(+l),s):t},s.defined=function(l){return arguments.length?(n=typeof l=="function"?l:Te(!!l),s):n},s.curve=function(l){return arguments.length?(a=l,r!=null&&(o=a(r)),s):a},s.context=function(l){return arguments.length?(l==null?r=o=null:o=a(r=l),s):r},s}function rc(e,t,n){var r=null,a=Te(!0),o=null,i=ad,s=null,l=Uy(u);e=typeof e=="function"?e:e===void 0?h8:Te(+e),t=typeof t=="function"?t:Te(t===void 0?0:+t),n=typeof n=="function"?n:n===void 0?v8:Te(+n);function u(c){var f,m,h,d=(c=Wy(c)).length,g,v=!1,y,x=new Array(d),P=new Array(d);for(o==null&&(s=i(y=l())),f=0;f<=d;++f){if(!(f<d&&a(g=c[f],f,c))===v)if(v=!v)m=f,s.areaStart(),s.lineStart();else{for(s.lineEnd(),s.lineStart(),h=f-1;h>=m;--h)s.point(x[h],P[h]);s.lineEnd(),s.areaEnd()}v&&(x[f]=+e(g,f,c),P[f]=+t(g,f,c),s.point(r?+r(g,f,c):x[f],n?+n(g,f,c):P[f]))}if(y)return s=null,y+""||null}function p(){return y8().defined(a).curve(i).context(o)}return u.x=function(c){return arguments.length?(e=typeof c=="function"?c:Te(+c),r=null,u):e},u.x0=function(c){return arguments.length?(e=typeof c=="function"?c:Te(+c),u):e},u.x1=function(c){return arguments.length?(r=c==null?null:typeof c=="function"?c:Te(+c),u):r},u.y=function(c){return arguments.length?(t=typeof c=="function"?c:Te(+c),n=null,u):t},u.y0=function(c){return arguments.length?(t=typeof c=="function"?c:Te(+c),u):t},u.y1=function(c){return arguments.length?(n=c==null?null:typeof c=="function"?c:Te(+c),u):n},u.lineX0=u.lineY0=function(){return p().x(e).y(t)},u.lineY1=function(){return p().x(e).y(n)},u.lineX1=function(){return p().x(r).y(t)},u.defined=function(c){return arguments.length?(a=typeof c=="function"?c:Te(!!c),u):a},u.curve=function(c){return arguments.length?(i=c,o!=null&&(s=i(o)),u):i},u.context=function(c){return arguments.length?(c==null?o=s=null:s=i(o=c),u):o},u}class g8{constructor(t,n){this._context=t,this._x=n}areaStart(){this._line=0}areaEnd(){this._line=NaN}lineStart(){this._point=0}lineEnd(){(this._line||this._line!==0&&this._point===1)&&this._context.closePath(),this._line=1-this._line}point(t,n){switch(t=+t,n=+n,this._point){case 0:{this._point=1,this._line?this._context.lineTo(t,n):this._context.moveTo(t,n);break}case 1:this._point=2;default:{this._x?this._context.bezierCurveTo(this._x0=(this._x0+t)/2,this._y0,this._x0,n,t,n):this._context.bezierCurveTo(this._x0,this._y0=(this._y0+n)/2,t,this._y0,t,n);break}}this._x0=t,this._y0=n}}function $R(e){return new g8(e,!0)}function MR(e){return new g8(e,!1)}const qy={draw(e,t){const n=qn(t/gp);e.moveTo(n,0),e.arc(0,0,n,0,rd)}},RR={draw(e,t){const n=qn(t/5)/2;e.moveTo(-3*n,-n),e.lineTo(-n,-n),e.lineTo(-n,-3*n),e.lineTo(n,-3*n),e.lineTo(n,-n),e.lineTo(3*n,-n),e.lineTo(3*n,n),e.lineTo(n,n),e.lineTo(n,3*n),e.lineTo(-n,3*n),e.lineTo(-n,n),e.lineTo(-3*n,n),e.closePath()}},x8=qn(1/3),IR=x8*2,DR={draw(e,t){const n=qn(t/IR),r=n*x8;e.moveTo(0,-n),e.lineTo(r,0),e.lineTo(0,n),e.lineTo(-r,0),e.closePath()}},LR={draw(e,t){const n=qn(t),r=-n/2;e.rect(r,r,n,n)}},FR=.8908130915292852,w8=yp(gp/10)/yp(7*gp/10),BR=yp(rd/10)*w8,zR=-f8(rd/10)*w8,HR={draw(e,t){const n=qn(t*FR),r=BR*n,a=zR*n;e.moveTo(0,-n),e.lineTo(r,a);for(let o=1;o<5;++o){const i=rd*o/5,s=f8(i),l=yp(i);e.lineTo(l*n,-s*n),e.lineTo(s*r-l*a,l*r+s*a)}e.closePath()}},xm=qn(3),GR={draw(e,t){const n=-qn(t/(xm*3));e.moveTo(0,n*2),e.lineTo(-xm*n,-n),e.lineTo(xm*n,-n),e.closePath()}},sn=-.5,ln=qn(3)/2,o0=1/qn(12),UR=(o0/2+1)*3,WR={draw(e,t){const n=qn(t/UR),r=n/2,a=n*o0,o=r,i=n*o0+n,s=-o,l=i;e.moveTo(r,a),e.lineTo(o,i),e.lineTo(s,l),e.lineTo(sn*r-ln*a,ln*r+sn*a),e.lineTo(sn*o-ln*i,ln*o+sn*i),e.lineTo(sn*s-ln*l,ln*s+sn*l),e.lineTo(sn*r+ln*a,sn*a-ln*r),e.lineTo(sn*o+ln*i,sn*i-ln*o),e.lineTo(sn*s+ln*l,sn*l-ln*s),e.closePath()}};function qR(e,t){let n=null,r=Uy(a);e=typeof e=="function"?e:Te(e||qy),t=typeof t=="function"?t:Te(t===void 0?64:+t);function a(){let o;if(n||(n=o=r()),e.apply(this,arguments).draw(n,+t.apply(this,arguments)),o)return n=null,o+""||null}return a.type=function(o){return arguments.length?(e=typeof o=="function"?o:Te(o),a):e},a.size=function(o){return arguments.length?(t=typeof o=="function"?o:Te(+o),a):t},a.context=function(o){return arguments.length?(n=o??null,a):n},a}function xp(){}function wp(e,t,n){e._context.bezierCurveTo((2*e._x0+e._x1)/3,(2*e._y0+e._y1)/3,(e._x0+2*e._x1)/3,(e._y0+2*e._y1)/3,(e._x0+4*e._x1+t)/6,(e._y0+4*e._y1+n)/6)}function b8(e){this._context=e}b8.prototype={areaStart:function(){this._line=0},areaEnd:function(){this._line=NaN},lineStart:function(){this._x0=this._x1=this._y0=this._y1=NaN,this._point=0},lineEnd:function(){switch(this._point){case 3:wp(this,this._x1,this._y1);case 2:this._context.lineTo(this._x1,this._y1);break}(this._line||this._line!==0&&this._point===1)&&this._context.closePath(),this._line=1-this._line},point:function(e,t){switch(e=+e,t=+t,this._point){case 0:this._point=1,this._line?this._context.lineTo(e,t):this._context.moveTo(e,t);break;case 1:this._point=2;break;case 2:this._point=3,this._context.lineTo((5*this._x0+this._x1)/6,(5*this._y0+this._y1)/6);default:wp(this,e,t);break}this._x0=this._x1,this._x1=e,this._y0=this._y1,this._y1=t}};function VR(e){return new b8(e)}function P8(e){this._context=e}P8.prototype={areaStart:xp,areaEnd:xp,lineStart:function(){this._x0=this._x1=this._x2=this._x3=this._x4=this._y0=this._y1=this._y2=this._y3=this._y4=NaN,this._point=0},lineEnd:function(){switch(this._point){case 1:{this._context.moveTo(this._x2,this._y2),this._context.closePath();break}case 2:{this._context.moveTo((this._x2+2*this._x3)/3,(this._y2+2*this._y3)/3),this._context.lineTo((this._x3+2*this._x2)/3,(this._y3+2*this._y2)/3),this._context.closePath();break}case 3:{this.point(this._x2,this._y2),this.point(this._x3,this._y3),this.point(this._x4,this._y4);break}}},point:function(e,t){switch(e=+e,t=+t,this._point){case 0:this._point=1,this._x2=e,this._y2=t;break;case 1:this._point=2,this._x3=e,this._y3=t;break;case 2:this._point=3,this._x4=e,this._y4=t,this._context.moveTo((this._x0+4*this._x1+e)/6,(this._y0+4*this._y1+t)/6);break;default:wp(this,e,t);break}this._x0=this._x1,this._x1=e,this._y0=this._y1,this._y1=t}};function KR(e){return new P8(e)}function S8(e){this._context=e}S8.prototype={areaStart:function(){this._line=0},areaEnd:function(){this._line=NaN},lineStart:function(){this._x0=this._x1=this._y0=this._y1=NaN,this._point=0},lineEnd:function(){(this._line||this._line!==0&&this._point===3)&&this._context.closePath(),this._line=1-this._line},point:function(e,t){switch(e=+e,t=+t,this._point){case 0:this._point=1;break;case 1:this._point=2;break;case 2:this._point=3;var n=(this._x0+4*this._x1+e)/6,r=(this._y0+4*this._y1+t)/6;this._line?this._context.lineTo(n,r):this._context.moveTo(n,r);break;case 3:this._point=4;default:wp(this,e,t);break}this._x0=this._x1,this._x1=e,this._y0=this._y1,this._y1=t}};function XR(e){return new S8(e)}function O8(e){this._context=e}O8.prototype={areaStart:xp,areaEnd:xp,lineStart:function(){this._point=0},lineEnd:function(){this._point&&this._context.closePath()},point:function(e,t){e=+e,t=+t,this._point?this._context.lineTo(e,t):(this._point=1,this._context.moveTo(e,t))}};function YR(e){return new O8(e)}function Xx(e){return e<0?-1:1}function Yx(e,t,n){var r=e._x1-e._x0,a=t-e._x1,o=(e._y1-e._y0)/(r||a<0&&-0),i=(n-e._y1)/(a||r<0&&-0),s=(o*a+i*r)/(r+a);return(Xx(o)+Xx(i))*Math.min(Math.abs(o),Math.abs(i),.5*Math.abs(s))||0}function Qx(e,t){var n=e._x1-e._x0;return n?(3*(e._y1-e._y0)/n-t)/2:t}function wm(e,t,n){var r=e._x0,a=e._y0,o=e._x1,i=e._y1,s=(o-r)/3;e._context.bezierCurveTo(r+s,a+s*t,o-s,i-s*n,o,i)}function bp(e){this._context=e}bp.prototype={areaStart:function(){this._line=0},areaEnd:function(){this._line=NaN},lineStart:function(){this._x0=this._x1=this._y0=this._y1=this._t0=NaN,this._point=0},lineEnd:function(){switch(this._point){case 2:this._context.lineTo(this._x1,this._y1);break;case 3:wm(this,this._t0,Qx(this,this._t0));break}(this._line||this._line!==0&&this._point===1)&&this._context.closePath(),this._line=1-this._line},point:function(e,t){var n=NaN;if(e=+e,t=+t,!(e===this._x1&&t===this._y1)){switch(this._point){case 0:this._point=1,this._line?this._context.lineTo(e,t):this._context.moveTo(e,t);break;case 1:this._point=2;break;case 2:this._point=3,wm(this,Qx(this,n=Yx(this,e,t)),n);break;default:wm(this,this._t0,n=Yx(this,e,t));break}this._x0=this._x1,this._x1=e,this._y0=this._y1,this._y1=t,this._t0=n}}};function k8(e){this._context=new C8(e)}(k8.prototype=Object.create(bp.prototype)).point=function(e,t){bp.prototype.point.call(this,t,e)};function C8(e){this._context=e}C8.prototype={moveTo:function(e,t){this._context.moveTo(t,e)},closePath:function(){this._context.closePath()},lineTo:function(e,t){this._context.lineTo(t,e)},bezierCurveTo:function(e,t,n,r,a,o){this._context.bezierCurveTo(t,e,r,n,o,a)}};function QR(e){return new bp(e)}function ZR(e){return new k8(e)}function _8(e){this._context=e}_8.prototype={areaStart:function(){this._line=0},areaEnd:function(){this._line=NaN},lineStart:function(){this._x=[],this._y=[]},lineEnd:function(){var e=this._x,t=this._y,n=e.length;if(n)if(this._line?this._context.lineTo(e[0],t[0]):this._context.moveTo(e[0],t[0]),n===2)this._context.lineTo(e[1],t[1]);else for(var r=Zx(e),a=Zx(t),o=0,i=1;i<n;++o,++i)this._context.bezierCurveTo(r[0][o],a[0][o],r[1][o],a[1][o],e[i],t[i]);(this._line||this._line!==0&&n===1)&&this._context.closePath(),this._line=1-this._line,this._x=this._y=null},point:function(e,t){this._x.push(+e),this._y.push(+t)}};function Zx(e){var t,n=e.length-1,r,a=new Array(n),o=new Array(n),i=new Array(n);for(a[0]=0,o[0]=2,i[0]=e[0]+2*e[1],t=1;t<n-1;++t)a[t]=1,o[t]=4,i[t]=4*e[t]+2*e[t+1];for(a[n-1]=2,o[n-1]=7,i[n-1]=8*e[n-1]+e[n],t=1;t<n;++t)r=a[t]/o[t-1],o[t]-=r,i[t]-=r*i[t-1];for(a[n-1]=i[n-1]/o[n-1],t=n-2;t>=0;--t)a[t]=(i[t]-a[t+1])/o[t];for(o[n-1]=(e[n]+a[n-1])/2,t=0;t<n-1;++t)o[t]=2*e[t+1]-a[t+1];return[a,o]}function JR(e){return new _8(e)}function od(e,t){this._context=e,this._t=t}od.prototype={areaStart:function(){this._line=0},areaEnd:function(){this._line=NaN},lineStart:function(){this._x=this._y=NaN,this._point=0},lineEnd:function(){0<this._t&&this._t<1&&this._point===2&&this._context.lineTo(this._x,this._y),(this._line||this._line!==0&&this._point===1)&&this._context.closePath(),this._line>=0&&(this._t=1-this._t,this._line=1-this._line)},point:function(e,t){switch(e=+e,t=+t,this._point){case 0:this._point=1,this._line?this._context.lineTo(e,t):this._context.moveTo(e,t);break;case 1:this._point=2;default:{if(this._t<=0)this._context.lineTo(this._x,t),this._context.lineTo(e,t);else{var n=this._x*(1-this._t)+e*this._t;this._context.lineTo(n,this._y),this._context.lineTo(n,t)}break}}this._x=e,this._y=t}};function eI(e){return new od(e,.5)}function tI(e){return new od(e,0)}function nI(e){return new od(e,1)}function ji(e,t){if((i=e.length)>1)for(var n=1,r,a,o=e[t[0]],i,s=o.length;n<i;++n)for(a=o,o=e[t[n]],r=0;r<s;++r)o[r][1]+=o[r][0]=isNaN(a[r][1])?a[r][0]:a[r][1]}function i0(e){for(var t=e.length,n=new Array(t);--t>=0;)n[t]=t;return n}function rI(e,t){return e[t]}function aI(e){const t=[];return t.key=e,t}function oI(){var e=Te([]),t=i0,n=ji,r=rI;function a(o){var i=Array.from(e.apply(this,arguments),aI),s,l=i.length,u=-1,p;for(const c of o)for(s=0,++u;s<l;++s)(i[s][u]=[0,+r(c,i[s].key,u,o)]).data=c;for(s=0,p=Wy(t(i));s<l;++s)i[p[s]].index=s;return n(i,p),i}return a.keys=function(o){return arguments.length?(e=typeof o=="function"?o:Te(Array.from(o)),a):e},a.value=function(o){return arguments.length?(r=typeof o=="function"?o:Te(+o),a):r},a.order=function(o){return arguments.length?(t=o==null?i0:typeof o=="function"?o:Te(Array.from(o)),a):t},a.offset=function(o){return arguments.length?(n=o??ji,a):n},a}function iI(e,t){if((r=e.length)>0){for(var n,r,a=0,o=e[0].length,i;a<o;++a){for(i=n=0;n<r;++n)i+=e[n][a][1]||0;if(i)for(n=0;n<r;++n)e[n][a][1]/=i}ji(e,t)}}function sI(e,t){if((a=e.length)>0){for(var n=0,r=e[t[0]],a,o=r.length;n<o;++n){for(var i=0,s=0;i<a;++i)s+=e[i][n][1]||0;r[n][1]+=r[n][0]=-s/2}ji(e,t)}}function lI(e,t){if(!(!((i=e.length)>0)||!((o=(a=e[t[0]]).length)>0))){for(var n=0,r=1,a,o,i;r<o;++r){for(var s=0,l=0,u=0;s<i;++s){for(var p=e[t[s]],c=p[r][1]||0,f=p[r-1][1]||0,m=(c-f)/2,h=0;h<s;++h){var d=e[t[h]],g=d[r][1]||0,v=d[r-1][1]||0;m+=g-v}l+=c,u+=m*c}a[r-1][1]+=a[r-1][0]=n,l&&(n-=u/l)}a[r-1][1]+=a[r-1][0]=n,ji(e,t)}}function jl(e){"@babel/helpers - typeof";return jl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},jl(e)}var uI=["type","size","sizeType"];function s0(){return s0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},s0.apply(this,arguments)}function Jx(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function ew(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Jx(Object(n),!0).forEach(function(r){cI(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Jx(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function cI(e,t,n){return t=pI(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function pI(e){var t=fI(e,"string");return jl(t)=="symbol"?t:t+""}function fI(e,t){if(jl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(jl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function dI(e,t){if(e==null)return{};var n=mI(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function mI(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}var A8={symbolCircle:qy,symbolCross:RR,symbolDiamond:DR,symbolSquare:LR,symbolStar:HR,symbolTriangle:GR,symbolWye:WR},hI=Math.PI/180,vI=function(t){var n="symbol".concat(nd(t));return A8[n]||qy},yI=function(t,n,r){if(n==="area")return t;switch(r){case"cross":return 5*t*t/9;case"diamond":return .5*t*t/Math.sqrt(3);case"square":return t*t;case"star":{var a=18*hI;return 1.25*t*t*(Math.tan(a)-Math.tan(a*2)*Math.pow(Math.tan(a),2))}case"triangle":return Math.sqrt(3)*t*t/4;case"wye":return(21-10*Math.sqrt(3))*t*t/8;default:return Math.PI*t*t/4}},gI=function(t,n){A8["symbol".concat(nd(t))]=n},Vy=function(t){var n=t.type,r=n===void 0?"circle":n,a=t.size,o=a===void 0?64:a,i=t.sizeType,s=i===void 0?"area":i,l=dI(t,uI),u=ew(ew({},l),{},{type:r,size:o,sizeType:s}),p=function(){var g=vI(r),v=qR().type(g).size(yI(o,s,r));return v()},c=u.className,f=u.cx,m=u.cy,h=ie(u,!0);return f===+f&&m===+m&&o===+o?E.createElement("path",s0({},h,{className:ue("recharts-symbols",c),transform:"translate(".concat(f,", ").concat(m,")"),d:p()})):null};Vy.registerSymbol=gI;function Ni(e){"@babel/helpers - typeof";return Ni=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Ni(e)}function l0(){return l0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},l0.apply(this,arguments)}function tw(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function xI(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?tw(Object(n),!0).forEach(function(r){Nl(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):tw(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function wI(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function bI(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,T8(r.key),r)}}function PI(e,t,n){return t&&bI(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function SI(e,t,n){return t=Pp(t),OI(e,E8()?Reflect.construct(t,n||[],Pp(e).constructor):t.apply(e,n))}function OI(e,t){if(t&&(Ni(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return kI(e)}function kI(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function E8(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(E8=function(){return!!e})()}function Pp(e){return Pp=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},Pp(e)}function CI(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&u0(e,t)}function u0(e,t){return u0=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},u0(e,t)}function Nl(e,t,n){return t=T8(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function T8(e){var t=_I(e,"string");return Ni(t)=="symbol"?t:t+""}function _I(e,t){if(Ni(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Ni(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var un=32,Ky=function(e){function t(){return wI(this,t),SI(this,t,arguments)}return CI(t,e),PI(t,[{key:"renderIcon",value:function(r){var a=this.props.inactiveColor,o=un/2,i=un/6,s=un/3,l=r.inactive?a:r.color;if(r.type==="plainline")return E.createElement("line",{strokeWidth:4,fill:"none",stroke:l,strokeDasharray:r.payload.strokeDasharray,x1:0,y1:o,x2:un,y2:o,className:"recharts-legend-icon"});if(r.type==="line")return E.createElement("path",{strokeWidth:4,fill:"none",stroke:l,d:"M0,".concat(o,"h").concat(s,`
            A`).concat(i,",").concat(i,",0,1,1,").concat(2*s,",").concat(o,`
            H`).concat(un,"M").concat(2*s,",").concat(o,`
            A`).concat(i,",").concat(i,",0,1,1,").concat(s,",").concat(o),className:"recharts-legend-icon"});if(r.type==="rect")return E.createElement("path",{stroke:"none",fill:l,d:"M0,".concat(un/8,"h").concat(un,"v").concat(un*3/4,"h").concat(-un,"z"),className:"recharts-legend-icon"});if(E.isValidElement(r.legendIcon)){var u=xI({},r);return delete u.legendIcon,E.cloneElement(r.legendIcon,u)}return E.createElement(Vy,{fill:l,cx:o,cy:o,size:un,sizeType:"diameter",type:r.type})}},{key:"renderItems",value:function(){var r=this,a=this.props,o=a.payload,i=a.iconSize,s=a.layout,l=a.formatter,u=a.inactiveColor,p={x:0,y:0,width:un,height:un},c={display:s==="horizontal"?"inline-block":"block",marginRight:10},f={display:"inline-block",verticalAlign:"middle",marginRight:4};return o.map(function(m,h){var d=m.formatter||l,g=ue(Nl(Nl({"recharts-legend-item":!0},"legend-item-".concat(h),!0),"inactive",m.inactive));if(m.type==="none")return null;var v=ae(m.value)?null:m.value;Tr(!ae(m.value),`The name property is also required when using a function for the dataKey of a chart's cartesian components. Ex: <Bar name="Name of my Data"/>`);var y=m.inactive?u:m.color;return E.createElement("li",l0({className:g,style:c,key:"legend-item-".concat(h)},vp(r.props,m,h)),E.createElement(Jh,{width:i,height:i,viewBox:p,style:f},r.renderIcon(m)),E.createElement("span",{className:"recharts-legend-item-text",style:{color:y}},d?d(v,m,h):v))})}},{key:"render",value:function(){var r=this.props,a=r.payload,o=r.layout,i=r.align;if(!a||!a.length)return null;var s={padding:0,margin:0,textAlign:o==="horizontal"?i:"left"};return E.createElement("ul",{className:"recharts-default-legend",style:s},this.renderItems())}}])}(k.PureComponent);Nl(Ky,"displayName","Legend");Nl(Ky,"defaultProps",{iconSize:14,layout:"horizontal",align:"center",verticalAlign:"middle",inactiveColor:"#ccc"});var AI=Gf;function EI(){this.__data__=new AI,this.size=0}var TI=EI;function jI(e){var t=this.__data__,n=t.delete(e);return this.size=t.size,n}var NI=jI;function $I(e){return this.__data__.get(e)}var MI=$I;function RI(e){return this.__data__.has(e)}var II=RI,DI=Gf,LI=Ry,FI=Iy,BI=200;function zI(e,t){var n=this.__data__;if(n instanceof DI){var r=n.__data__;if(!LI||r.length<BI-1)return r.push([e,t]),this.size=++n.size,this;n=this.__data__=new FI(r)}return n.set(e,t),this.size=n.size,this}var HI=zI,GI=Gf,UI=TI,WI=NI,qI=MI,VI=II,KI=HI;function gs(e){var t=this.__data__=new GI(e);this.size=t.size}gs.prototype.clear=UI;gs.prototype.delete=WI;gs.prototype.get=qI;gs.prototype.has=VI;gs.prototype.set=KI;var j8=gs,XI="__lodash_hash_undefined__";function YI(e){return this.__data__.set(e,XI),this}var QI=YI;function ZI(e){return this.__data__.has(e)}var JI=ZI,eD=Iy,tD=QI,nD=JI;function Sp(e){var t=-1,n=e==null?0:e.length;for(this.__data__=new eD;++t<n;)this.add(e[t])}Sp.prototype.add=Sp.prototype.push=tD;Sp.prototype.has=nD;var N8=Sp;function rD(e,t){for(var n=-1,r=e==null?0:e.length;++n<r;)if(t(e[n],n,e))return!0;return!1}var $8=rD;function aD(e,t){return e.has(t)}var M8=aD,oD=N8,iD=$8,sD=M8,lD=1,uD=2;function cD(e,t,n,r,a,o){var i=n&lD,s=e.length,l=t.length;if(s!=l&&!(i&&l>s))return!1;var u=o.get(e),p=o.get(t);if(u&&p)return u==t&&p==e;var c=-1,f=!0,m=n&uD?new oD:void 0;for(o.set(e,t),o.set(t,e);++c<s;){var h=e[c],d=t[c];if(r)var g=i?r(d,h,c,t,e,o):r(h,d,c,e,t,o);if(g!==void 0){if(g)continue;f=!1;break}if(m){if(!iD(t,function(v,y){if(!sD(m,y)&&(h===v||a(h,v,n,r,o)))return m.push(y)})){f=!1;break}}else if(!(h===d||a(h,d,n,r,o))){f=!1;break}}return o.delete(e),o.delete(t),f}var R8=cD,pD=vr,fD=pD.Uint8Array,dD=fD;function mD(e){var t=-1,n=Array(e.size);return e.forEach(function(r,a){n[++t]=[a,r]}),n}var hD=mD;function vD(e){var t=-1,n=Array(e.size);return e.forEach(function(r){n[++t]=r}),n}var Xy=vD,nw=Ou,rw=dD,yD=My,gD=R8,xD=hD,wD=Xy,bD=1,PD=2,SD="[object Boolean]",OD="[object Date]",kD="[object Error]",CD="[object Map]",_D="[object Number]",AD="[object RegExp]",ED="[object Set]",TD="[object String]",jD="[object Symbol]",ND="[object ArrayBuffer]",$D="[object DataView]",aw=nw?nw.prototype:void 0,bm=aw?aw.valueOf:void 0;function MD(e,t,n,r,a,o,i){switch(n){case $D:if(e.byteLength!=t.byteLength||e.byteOffset!=t.byteOffset)return!1;e=e.buffer,t=t.buffer;case ND:return!(e.byteLength!=t.byteLength||!o(new rw(e),new rw(t)));case SD:case OD:case _D:return yD(+e,+t);case kD:return e.name==t.name&&e.message==t.message;case AD:case TD:return e==t+"";case CD:var s=xD;case ED:var l=r&bD;if(s||(s=wD),e.size!=t.size&&!l)return!1;var u=i.get(e);if(u)return u==t;r|=PD,i.set(e,t);var p=gD(s(e),s(t),r,a,o,i);return i.delete(e),p;case jD:if(bm)return bm.call(e)==bm.call(t)}return!1}var RD=MD;function ID(e,t){for(var n=-1,r=t.length,a=e.length;++n<r;)e[a+n]=t[n];return e}var I8=ID,DD=I8,LD=qt;function FD(e,t,n){var r=t(e);return LD(e)?r:DD(r,n(e))}var BD=FD;function zD(e,t){for(var n=-1,r=e==null?0:e.length,a=0,o=[];++n<r;){var i=e[n];t(i,n,e)&&(o[a++]=i)}return o}var HD=zD;function GD(){return[]}var UD=GD,WD=HD,qD=UD,VD=Object.prototype,KD=VD.propertyIsEnumerable,ow=Object.getOwnPropertySymbols,XD=ow?function(e){return e==null?[]:(e=Object(e),WD(ow(e),function(t){return KD.call(e,t)}))}:qD,YD=XD;function QD(e,t){for(var n=-1,r=Array(e);++n<e;)r[n]=t(n);return r}var ZD=QD,JD=Gr,eL=Ur,tL="[object Arguments]";function nL(e){return eL(e)&&JD(e)==tL}var rL=nL,iw=rL,aL=Ur,D8=Object.prototype,oL=D8.hasOwnProperty,iL=D8.propertyIsEnumerable,sL=iw(function(){return arguments}())?iw:function(e){return aL(e)&&oL.call(e,"callee")&&!iL.call(e,"callee")},Yy=sL,Op={exports:{}};function lL(){return!1}var uL=lL;Op.exports;(function(e,t){var n=vr,r=uL,a=t&&!t.nodeType&&t,o=a&&!0&&e&&!e.nodeType&&e,i=o&&o.exports===a,s=i?n.Buffer:void 0,l=s?s.isBuffer:void 0,u=l||r;e.exports=u})(Op,Op.exports);var L8=Op.exports,cL=9007199254740991,pL=/^(?:0|[1-9]\d*)$/;function fL(e,t){var n=typeof e;return t=t??cL,!!t&&(n=="number"||n!="symbol"&&pL.test(e))&&e>-1&&e%1==0&&e<t}var Qy=fL,dL=9007199254740991;function mL(e){return typeof e=="number"&&e>-1&&e%1==0&&e<=dL}var Zy=mL,hL=Gr,vL=Zy,yL=Ur,gL="[object Arguments]",xL="[object Array]",wL="[object Boolean]",bL="[object Date]",PL="[object Error]",SL="[object Function]",OL="[object Map]",kL="[object Number]",CL="[object Object]",_L="[object RegExp]",AL="[object Set]",EL="[object String]",TL="[object WeakMap]",jL="[object ArrayBuffer]",NL="[object DataView]",$L="[object Float32Array]",ML="[object Float64Array]",RL="[object Int8Array]",IL="[object Int16Array]",DL="[object Int32Array]",LL="[object Uint8Array]",FL="[object Uint8ClampedArray]",BL="[object Uint16Array]",zL="[object Uint32Array]",De={};De[$L]=De[ML]=De[RL]=De[IL]=De[DL]=De[LL]=De[FL]=De[BL]=De[zL]=!0;De[gL]=De[xL]=De[jL]=De[wL]=De[NL]=De[bL]=De[PL]=De[SL]=De[OL]=De[kL]=De[CL]=De[_L]=De[AL]=De[EL]=De[TL]=!1;function HL(e){return yL(e)&&vL(e.length)&&!!De[hL(e)]}var GL=HL;function UL(e){return function(t){return e(t)}}var F8=UL,kp={exports:{}};kp.exports;(function(e,t){var n=q4,r=t&&!t.nodeType&&t,a=r&&!0&&e&&!e.nodeType&&e,o=a&&a.exports===r,i=o&&n.process,s=function(){try{var l=a&&a.require&&a.require("util").types;return l||i&&i.binding&&i.binding("util")}catch{}}();e.exports=s})(kp,kp.exports);var WL=kp.exports,qL=GL,VL=F8,sw=WL,lw=sw&&sw.isTypedArray,KL=lw?VL(lw):qL,B8=KL,XL=ZD,YL=Yy,QL=qt,ZL=L8,JL=Qy,eF=B8,tF=Object.prototype,nF=tF.hasOwnProperty;function rF(e,t){var n=QL(e),r=!n&&YL(e),a=!n&&!r&&ZL(e),o=!n&&!r&&!a&&eF(e),i=n||r||a||o,s=i?XL(e.length,String):[],l=s.length;for(var u in e)(t||nF.call(e,u))&&!(i&&(u=="length"||a&&(u=="offset"||u=="parent")||o&&(u=="buffer"||u=="byteLength"||u=="byteOffset")||JL(u,l)))&&s.push(u);return s}var aF=rF,oF=Object.prototype;function iF(e){var t=e&&e.constructor,n=typeof t=="function"&&t.prototype||oF;return e===n}var sF=iF;function lF(e,t){return function(n){return e(t(n))}}var z8=lF,uF=z8,cF=uF(Object.keys,Object),pF=cF,fF=sF,dF=pF,mF=Object.prototype,hF=mF.hasOwnProperty;function vF(e){if(!fF(e))return dF(e);var t=[];for(var n in Object(e))hF.call(e,n)&&n!="constructor"&&t.push(n);return t}var yF=vF,gF=$y,xF=Zy;function wF(e){return e!=null&&xF(e.length)&&!gF(e)}var ku=wF,bF=aF,PF=yF,SF=ku;function OF(e){return SF(e)?bF(e):PF(e)}var id=OF,kF=BD,CF=YD,_F=id;function AF(e){return kF(e,_F,CF)}var EF=AF,uw=EF,TF=1,jF=Object.prototype,NF=jF.hasOwnProperty;function $F(e,t,n,r,a,o){var i=n&TF,s=uw(e),l=s.length,u=uw(t),p=u.length;if(l!=p&&!i)return!1;for(var c=l;c--;){var f=s[c];if(!(i?f in t:NF.call(t,f)))return!1}var m=o.get(e),h=o.get(t);if(m&&h)return m==t&&h==e;var d=!0;o.set(e,t),o.set(t,e);for(var g=i;++c<l;){f=s[c];var v=e[f],y=t[f];if(r)var x=i?r(y,v,f,t,e,o):r(v,y,f,e,t,o);if(!(x===void 0?v===y||a(v,y,n,r,o):x)){d=!1;break}g||(g=f=="constructor")}if(d&&!g){var P=e.constructor,w=t.constructor;P!=w&&"constructor"in e&&"constructor"in t&&!(typeof P=="function"&&P instanceof P&&typeof w=="function"&&w instanceof w)&&(d=!1)}return o.delete(e),o.delete(t),d}var MF=$F,RF=Co,IF=vr,DF=RF(IF,"DataView"),LF=DF,FF=Co,BF=vr,zF=FF(BF,"Promise"),HF=zF,GF=Co,UF=vr,WF=GF(UF,"Set"),H8=WF,qF=Co,VF=vr,KF=qF(VF,"WeakMap"),XF=KF,c0=LF,p0=Ry,f0=HF,d0=H8,m0=XF,G8=Gr,xs=K4,cw="[object Map]",YF="[object Object]",pw="[object Promise]",fw="[object Set]",dw="[object WeakMap]",mw="[object DataView]",QF=xs(c0),ZF=xs(p0),JF=xs(f0),eB=xs(d0),tB=xs(m0),Ua=G8;(c0&&Ua(new c0(new ArrayBuffer(1)))!=mw||p0&&Ua(new p0)!=cw||f0&&Ua(f0.resolve())!=pw||d0&&Ua(new d0)!=fw||m0&&Ua(new m0)!=dw)&&(Ua=function(e){var t=G8(e),n=t==YF?e.constructor:void 0,r=n?xs(n):"";if(r)switch(r){case QF:return mw;case ZF:return cw;case JF:return pw;case eB:return fw;case tB:return dw}return t});var nB=Ua,Pm=j8,rB=R8,aB=RD,oB=MF,hw=nB,vw=qt,yw=L8,iB=B8,sB=1,gw="[object Arguments]",xw="[object Array]",ac="[object Object]",lB=Object.prototype,ww=lB.hasOwnProperty;function uB(e,t,n,r,a,o){var i=vw(e),s=vw(t),l=i?xw:hw(e),u=s?xw:hw(t);l=l==gw?ac:l,u=u==gw?ac:u;var p=l==ac,c=u==ac,f=l==u;if(f&&yw(e)){if(!yw(t))return!1;i=!0,p=!1}if(f&&!p)return o||(o=new Pm),i||iB(e)?rB(e,t,n,r,a,o):aB(e,t,l,n,r,a,o);if(!(n&sB)){var m=p&&ww.call(e,"__wrapped__"),h=c&&ww.call(t,"__wrapped__");if(m||h){var d=m?e.value():e,g=h?t.value():t;return o||(o=new Pm),a(d,g,n,r,o)}}return f?(o||(o=new Pm),oB(e,t,n,r,a,o)):!1}var cB=uB,pB=cB,bw=Ur;function U8(e,t,n,r,a){return e===t?!0:e==null||t==null||!bw(e)&&!bw(t)?e!==e&&t!==t:pB(e,t,n,r,U8,a)}var Jy=U8,fB=j8,dB=Jy,mB=1,hB=2;function vB(e,t,n,r){var a=n.length,o=a,i=!r;if(e==null)return!o;for(e=Object(e);a--;){var s=n[a];if(i&&s[2]?s[1]!==e[s[0]]:!(s[0]in e))return!1}for(;++a<o;){s=n[a];var l=s[0],u=e[l],p=s[1];if(i&&s[2]){if(u===void 0&&!(l in e))return!1}else{var c=new fB;if(r)var f=r(u,p,l,e,t,c);if(!(f===void 0?dB(p,u,mB|hB,r,c):f))return!1}}return!0}var yB=vB,gB=Na;function xB(e){return e===e&&!gB(e)}var W8=xB,wB=W8,bB=id;function PB(e){for(var t=bB(e),n=t.length;n--;){var r=t[n],a=e[r];t[n]=[r,a,wB(a)]}return t}var SB=PB;function OB(e,t){return function(n){return n==null?!1:n[e]===t&&(t!==void 0||e in Object(n))}}var q8=OB,kB=yB,CB=SB,_B=q8;function AB(e){var t=CB(e);return t.length==1&&t[0][2]?_B(t[0][0],t[0][1]):function(n){return n===e||kB(n,e,t)}}var EB=AB;function TB(e,t){return e!=null&&t in Object(e)}var jB=TB,NB=J4,$B=Yy,MB=qt,RB=Qy,IB=Zy,DB=Wf;function LB(e,t,n){t=NB(t,e);for(var r=-1,a=t.length,o=!1;++r<a;){var i=DB(t[r]);if(!(o=e!=null&&n(e,i)))break;e=e[i]}return o||++r!=a?o:(a=e==null?0:e.length,!!a&&IB(a)&&RB(i,a)&&(MB(e)||$B(e)))}var FB=LB,BB=jB,zB=FB;function HB(e,t){return e!=null&&zB(e,t,BB)}var GB=HB,UB=Jy,WB=e8,qB=GB,VB=Ny,KB=W8,XB=q8,YB=Wf,QB=1,ZB=2;function JB(e,t){return VB(e)&&KB(t)?XB(YB(e),t):function(n){var r=WB(n,e);return r===void 0&&r===t?qB(n,e):UB(t,r,QB|ZB)}}var ez=JB;function tz(e){return e}var ws=tz;function nz(e){return function(t){return t==null?void 0:t[e]}}var rz=nz,az=Fy;function oz(e){return function(t){return az(t,e)}}var iz=oz,sz=rz,lz=iz,uz=Ny,cz=Wf;function pz(e){return uz(e)?sz(cz(e)):lz(e)}var fz=pz,dz=EB,mz=ez,hz=ws,vz=qt,yz=fz;function gz(e){return typeof e=="function"?e:e==null?hz:typeof e=="object"?vz(e)?mz(e[0],e[1]):dz(e):yz(e)}var $a=gz;function xz(e,t,n,r){for(var a=e.length,o=n+(r?1:-1);r?o--:++o<a;)if(t(e[o],o,e))return o;return-1}var V8=xz;function wz(e){return e!==e}var bz=wz;function Pz(e,t,n){for(var r=n-1,a=e.length;++r<a;)if(e[r]===t)return r;return-1}var Sz=Pz,Oz=V8,kz=bz,Cz=Sz;function _z(e,t,n){return t===t?Cz(e,t,n):Oz(e,kz,n)}var Az=_z,Ez=Az;function Tz(e,t){var n=e==null?0:e.length;return!!n&&Ez(e,t,0)>-1}var jz=Tz;function Nz(e,t,n){for(var r=-1,a=e==null?0:e.length;++r<a;)if(n(t,e[r]))return!0;return!1}var $z=Nz;function Mz(){}var Rz=Mz,Sm=H8,Iz=Rz,Dz=Xy,Lz=1/0,Fz=Sm&&1/Dz(new Sm([,-0]))[1]==Lz?function(e){return new Sm(e)}:Iz,Bz=Fz,zz=N8,Hz=jz,Gz=$z,Uz=M8,Wz=Bz,qz=Xy,Vz=200;function Kz(e,t,n){var r=-1,a=Hz,o=e.length,i=!0,s=[],l=s;if(n)i=!1,a=Gz;else if(o>=Vz){var u=t?null:Wz(e);if(u)return qz(u);i=!1,a=Uz,l=new zz}else l=t?[]:s;e:for(;++r<o;){var p=e[r],c=t?t(p):p;if(p=n||p!==0?p:0,i&&c===c){for(var f=l.length;f--;)if(l[f]===c)continue e;t&&l.push(c),s.push(p)}else a(l,c,n)||(l!==s&&l.push(c),s.push(p))}return s}var Xz=Kz,Yz=$a,Qz=Xz;function Zz(e,t){return e&&e.length?Qz(e,Yz(t)):[]}var Jz=Zz;const Pw=_e(Jz);function K8(e,t,n){return t===!0?Pw(e,n):ae(t)?Pw(e,t):e}function $i(e){"@babel/helpers - typeof";return $i=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},$i(e)}var eH=["ref"];function Sw(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function gr(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Sw(Object(n),!0).forEach(function(r){sd(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Sw(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function tH(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Ow(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,Y8(r.key),r)}}function nH(e,t,n){return t&&Ow(e.prototype,t),n&&Ow(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function rH(e,t,n){return t=Cp(t),aH(e,X8()?Reflect.construct(t,n||[],Cp(e).constructor):t.apply(e,n))}function aH(e,t){if(t&&($i(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return oH(e)}function oH(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function X8(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(X8=function(){return!!e})()}function Cp(e){return Cp=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},Cp(e)}function iH(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&h0(e,t)}function h0(e,t){return h0=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},h0(e,t)}function sd(e,t,n){return t=Y8(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function Y8(e){var t=sH(e,"string");return $i(t)=="symbol"?t:t+""}function sH(e,t){if($i(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if($i(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}function lH(e,t){if(e==null)return{};var n=uH(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function uH(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function cH(e){return e.value}function pH(e,t){if(E.isValidElement(e))return E.cloneElement(e,t);if(typeof e=="function")return E.createElement(e,t);t.ref;var n=lH(t,eH);return E.createElement(Ky,n)}var kw=1,ci=function(e){function t(){var n;tH(this,t);for(var r=arguments.length,a=new Array(r),o=0;o<r;o++)a[o]=arguments[o];return n=rH(this,t,[].concat(a)),sd(n,"lastBoundingBox",{width:-1,height:-1}),n}return iH(t,e),nH(t,[{key:"componentDidMount",value:function(){this.updateBBox()}},{key:"componentDidUpdate",value:function(){this.updateBBox()}},{key:"getBBox",value:function(){if(this.wrapperNode&&this.wrapperNode.getBoundingClientRect){var r=this.wrapperNode.getBoundingClientRect();return r.height=this.wrapperNode.offsetHeight,r.width=this.wrapperNode.offsetWidth,r}return null}},{key:"updateBBox",value:function(){var r=this.props.onBBoxUpdate,a=this.getBBox();a?(Math.abs(a.width-this.lastBoundingBox.width)>kw||Math.abs(a.height-this.lastBoundingBox.height)>kw)&&(this.lastBoundingBox.width=a.width,this.lastBoundingBox.height=a.height,r&&r(a)):(this.lastBoundingBox.width!==-1||this.lastBoundingBox.height!==-1)&&(this.lastBoundingBox.width=-1,this.lastBoundingBox.height=-1,r&&r(null))}},{key:"getBBoxSnapshot",value:function(){return this.lastBoundingBox.width>=0&&this.lastBoundingBox.height>=0?gr({},this.lastBoundingBox):{width:0,height:0}}},{key:"getDefaultPosition",value:function(r){var a=this.props,o=a.layout,i=a.align,s=a.verticalAlign,l=a.margin,u=a.chartWidth,p=a.chartHeight,c,f;if(!r||(r.left===void 0||r.left===null)&&(r.right===void 0||r.right===null))if(i==="center"&&o==="vertical"){var m=this.getBBoxSnapshot();c={left:((u||0)-m.width)/2}}else c=i==="right"?{right:l&&l.right||0}:{left:l&&l.left||0};if(!r||(r.top===void 0||r.top===null)&&(r.bottom===void 0||r.bottom===null))if(s==="middle"){var h=this.getBBoxSnapshot();f={top:((p||0)-h.height)/2}}else f=s==="bottom"?{bottom:l&&l.bottom||0}:{top:l&&l.top||0};return gr(gr({},c),f)}},{key:"render",value:function(){var r=this,a=this.props,o=a.content,i=a.width,s=a.height,l=a.wrapperStyle,u=a.payloadUniqBy,p=a.payload,c=gr(gr({position:"absolute",width:i||"auto",height:s||"auto"},this.getDefaultPosition(l)),l);return E.createElement("div",{className:"recharts-legend-wrapper",style:c,ref:function(m){r.wrapperNode=m}},pH(o,gr(gr({},this.props),{},{payload:K8(p,u,cH)})))}}],[{key:"getWithHeight",value:function(r,a){var o=gr(gr({},this.defaultProps),r.props),i=o.layout;return i==="vertical"&&V(r.props.height)?{height:r.props.height}:i==="horizontal"?{width:r.props.width||a}:null}}])}(k.PureComponent);sd(ci,"displayName","Legend");sd(ci,"defaultProps",{iconSize:14,layout:"horizontal",align:"center",verticalAlign:"bottom"});var Cw=Ou,fH=Yy,dH=qt,_w=Cw?Cw.isConcatSpreadable:void 0;function mH(e){return dH(e)||fH(e)||!!(_w&&e&&e[_w])}var hH=mH,vH=I8,yH=hH;function Q8(e,t,n,r,a){var o=-1,i=e.length;for(n||(n=yH),a||(a=[]);++o<i;){var s=e[o];t>0&&n(s)?t>1?Q8(s,t-1,n,r,a):vH(a,s):r||(a[a.length]=s)}return a}var Z8=Q8;function gH(e){return function(t,n,r){for(var a=-1,o=Object(t),i=r(t),s=i.length;s--;){var l=i[e?s:++a];if(n(o[l],l,o)===!1)break}return t}}var xH=gH,wH=xH,bH=wH(),PH=bH,SH=PH,OH=id;function kH(e,t){return e&&SH(e,t,OH)}var J8=kH,CH=ku;function _H(e,t){return function(n,r){if(n==null)return n;if(!CH(n))return e(n,r);for(var a=n.length,o=t?a:-1,i=Object(n);(t?o--:++o<a)&&r(i[o],o,i)!==!1;);return n}}var AH=_H,EH=J8,TH=AH,jH=TH(EH),eg=jH,NH=eg,$H=ku;function MH(e,t){var n=-1,r=$H(e)?Array(e.length):[];return NH(e,function(a,o,i){r[++n]=t(a,o,i)}),r}var e7=MH;function RH(e,t){var n=e.length;for(e.sort(t);n--;)e[n]=e[n].value;return e}var IH=RH,Aw=ps;function DH(e,t){if(e!==t){var n=e!==void 0,r=e===null,a=e===e,o=Aw(e),i=t!==void 0,s=t===null,l=t===t,u=Aw(t);if(!s&&!u&&!o&&e>t||o&&i&&l&&!s&&!u||r&&i&&l||!n&&l||!a)return 1;if(!r&&!o&&!u&&e<t||u&&n&&a&&!r&&!o||s&&n&&a||!i&&a||!l)return-1}return 0}var LH=DH,FH=LH;function BH(e,t,n){for(var r=-1,a=e.criteria,o=t.criteria,i=a.length,s=n.length;++r<i;){var l=FH(a[r],o[r]);if(l){if(r>=s)return l;var u=n[r];return l*(u=="desc"?-1:1)}}return e.index-t.index}var zH=BH,Om=Ly,HH=Fy,GH=$a,UH=e7,WH=IH,qH=F8,VH=zH,KH=ws,XH=qt;function YH(e,t,n){t.length?t=Om(t,function(o){return XH(o)?function(i){return HH(i,o.length===1?o[0]:o)}:o}):t=[KH];var r=-1;t=Om(t,qH(GH));var a=UH(e,function(o,i,s){var l=Om(t,function(u){return u(o)});return{criteria:l,index:++r,value:o}});return WH(a,function(o,i){return VH(o,i,n)})}var QH=YH;function ZH(e,t,n){switch(n.length){case 0:return e.call(t);case 1:return e.call(t,n[0]);case 2:return e.call(t,n[0],n[1]);case 3:return e.call(t,n[0],n[1],n[2])}return e.apply(t,n)}var JH=ZH,eG=JH,Ew=Math.max;function tG(e,t,n){return t=Ew(t===void 0?e.length-1:t,0),function(){for(var r=arguments,a=-1,o=Ew(r.length-t,0),i=Array(o);++a<o;)i[a]=r[t+a];a=-1;for(var s=Array(t+1);++a<t;)s[a]=r[a];return s[t]=n(i),eG(e,this,s)}}var nG=tG;function rG(e){return function(){return e}}var aG=rG,oG=Co,iG=function(){try{var e=oG(Object,"defineProperty");return e({},"",{}),e}catch{}}(),t7=iG,sG=aG,Tw=t7,lG=ws,uG=Tw?function(e,t){return Tw(e,"toString",{configurable:!0,enumerable:!1,value:sG(t),writable:!0})}:lG,cG=uG,pG=800,fG=16,dG=Date.now;function mG(e){var t=0,n=0;return function(){var r=dG(),a=fG-(r-n);if(n=r,a>0){if(++t>=pG)return arguments[0]}else t=0;return e.apply(void 0,arguments)}}var hG=mG,vG=cG,yG=hG,gG=yG(vG),xG=gG,wG=ws,bG=nG,PG=xG;function SG(e,t){return PG(bG(e,t,wG),e+"")}var OG=SG,kG=My,CG=ku,_G=Qy,AG=Na;function EG(e,t,n){if(!AG(n))return!1;var r=typeof t;return(r=="number"?CG(n)&&_G(t,n.length):r=="string"&&t in n)?kG(n[t],e):!1}var ld=EG,TG=Z8,jG=QH,NG=OG,jw=ld,$G=NG(function(e,t){if(e==null)return[];var n=t.length;return n>1&&jw(e,t[0],t[1])?t=[]:n>2&&jw(t[0],t[1],t[2])&&(t=[t[0]]),jG(e,TG(t,1),[])}),MG=$G;const tg=_e(MG);function $l(e){"@babel/helpers - typeof";return $l=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},$l(e)}function v0(){return v0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},v0.apply(this,arguments)}function RG(e,t){return FG(e)||LG(e,t)||DG(e,t)||IG()}function IG(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function DG(e,t){if(e){if(typeof e=="string")return Nw(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return Nw(e,t)}}function Nw(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function LG(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function FG(e){if(Array.isArray(e))return e}function $w(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function km(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?$w(Object(n),!0).forEach(function(r){BG(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):$w(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function BG(e,t,n){return t=zG(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function zG(e){var t=HG(e,"string");return $l(t)=="symbol"?t:t+""}function HG(e,t){if($l(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if($l(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function GG(e){return Array.isArray(e)&&ot(e[0])&&ot(e[1])?e.join(" ~ "):e}var UG=function(t){var n=t.separator,r=n===void 0?" : ":n,a=t.contentStyle,o=a===void 0?{}:a,i=t.itemStyle,s=i===void 0?{}:i,l=t.labelStyle,u=l===void 0?{}:l,p=t.payload,c=t.formatter,f=t.itemSorter,m=t.wrapperClassName,h=t.labelClassName,d=t.label,g=t.labelFormatter,v=t.accessibilityLayer,y=v===void 0?!1:v,x=function(){if(p&&p.length){var j={padding:0,margin:0},$=(f?tg(p,f):p).map(function(M,I){if(M.type==="none")return null;var R=km({display:"block",paddingTop:4,paddingBottom:4,color:M.color||"#000"},s),L=M.formatter||c||GG,N=M.value,D=M.name,H=N,W=D;if(L&&H!=null&&W!=null){var G=L(N,D,M,I,p);if(Array.isArray(G)){var Z=RG(G,2);H=Z[0],W=Z[1]}else H=G}return E.createElement("li",{className:"recharts-tooltip-item",key:"tooltip-item-".concat(I),style:R},ot(W)?E.createElement("span",{className:"recharts-tooltip-item-name"},W):null,ot(W)?E.createElement("span",{className:"recharts-tooltip-item-separator"},r):null,E.createElement("span",{className:"recharts-tooltip-item-value"},H),E.createElement("span",{className:"recharts-tooltip-item-unit"},M.unit||""))});return E.createElement("ul",{className:"recharts-tooltip-item-list",style:j},$)}return null},P=km({margin:0,padding:10,backgroundColor:"#fff",border:"1px solid #ccc",whiteSpace:"nowrap"},o),w=km({margin:0},u),S=!le(d),O=S?d:"",C=ue("recharts-default-tooltip",m),_=ue("recharts-tooltip-label",h);S&&g&&p!==void 0&&p!==null&&(O=g(d,p));var T=y?{role:"status","aria-live":"assertive"}:{};return E.createElement("div",v0({className:C,style:P},T),E.createElement("p",{className:_,style:w},E.isValidElement(O)?O:"".concat(O)),x())};function Ml(e){"@babel/helpers - typeof";return Ml=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Ml(e)}function oc(e,t,n){return t=WG(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function WG(e){var t=qG(e,"string");return Ml(t)=="symbol"?t:t+""}function qG(e,t){if(Ml(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Ml(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var Is="recharts-tooltip-wrapper",VG={visibility:"hidden"};function KG(e){var t=e.coordinate,n=e.translateX,r=e.translateY;return ue(Is,oc(oc(oc(oc({},"".concat(Is,"-right"),V(n)&&t&&V(t.x)&&n>=t.x),"".concat(Is,"-left"),V(n)&&t&&V(t.x)&&n<t.x),"".concat(Is,"-bottom"),V(r)&&t&&V(t.y)&&r>=t.y),"".concat(Is,"-top"),V(r)&&t&&V(t.y)&&r<t.y))}function Mw(e){var t=e.allowEscapeViewBox,n=e.coordinate,r=e.key,a=e.offsetTopLeft,o=e.position,i=e.reverseDirection,s=e.tooltipDimension,l=e.viewBox,u=e.viewBoxDimension;if(o&&V(o[r]))return o[r];var p=n[r]-s-a,c=n[r]+a;if(t[r])return i[r]?p:c;if(i[r]){var f=p,m=l[r];return f<m?Math.max(c,l[r]):Math.max(p,l[r])}var h=c+s,d=l[r]+u;return h>d?Math.max(p,l[r]):Math.max(c,l[r])}function XG(e){var t=e.translateX,n=e.translateY,r=e.useTranslate3d;return{transform:r?"translate3d(".concat(t,"px, ").concat(n,"px, 0)"):"translate(".concat(t,"px, ").concat(n,"px)")}}function YG(e){var t=e.allowEscapeViewBox,n=e.coordinate,r=e.offsetTopLeft,a=e.position,o=e.reverseDirection,i=e.tooltipBox,s=e.useTranslate3d,l=e.viewBox,u,p,c;return i.height>0&&i.width>0&&n?(p=Mw({allowEscapeViewBox:t,coordinate:n,key:"x",offsetTopLeft:r,position:a,reverseDirection:o,tooltipDimension:i.width,viewBox:l,viewBoxDimension:l.width}),c=Mw({allowEscapeViewBox:t,coordinate:n,key:"y",offsetTopLeft:r,position:a,reverseDirection:o,tooltipDimension:i.height,viewBox:l,viewBoxDimension:l.height}),u=XG({translateX:p,translateY:c,useTranslate3d:s})):u=VG,{cssProperties:u,cssClasses:KG({translateX:p,translateY:c,coordinate:n})}}function Mi(e){"@babel/helpers - typeof";return Mi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Mi(e)}function Rw(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Iw(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Rw(Object(n),!0).forEach(function(r){g0(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Rw(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function QG(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function ZG(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,r7(r.key),r)}}function JG(e,t,n){return t&&ZG(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function eU(e,t,n){return t=_p(t),tU(e,n7()?Reflect.construct(t,n||[],_p(e).constructor):t.apply(e,n))}function tU(e,t){if(t&&(Mi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return nU(e)}function nU(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function n7(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(n7=function(){return!!e})()}function _p(e){return _p=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},_p(e)}function rU(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&y0(e,t)}function y0(e,t){return y0=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},y0(e,t)}function g0(e,t,n){return t=r7(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function r7(e){var t=aU(e,"string");return Mi(t)=="symbol"?t:t+""}function aU(e,t){if(Mi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Mi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var Dw=1,oU=function(e){function t(){var n;QG(this,t);for(var r=arguments.length,a=new Array(r),o=0;o<r;o++)a[o]=arguments[o];return n=eU(this,t,[].concat(a)),g0(n,"state",{dismissed:!1,dismissedAtCoordinate:{x:0,y:0},lastBoundingBox:{width:-1,height:-1}}),g0(n,"handleKeyDown",function(i){if(i.key==="Escape"){var s,l,u,p;n.setState({dismissed:!0,dismissedAtCoordinate:{x:(s=(l=n.props.coordinate)===null||l===void 0?void 0:l.x)!==null&&s!==void 0?s:0,y:(u=(p=n.props.coordinate)===null||p===void 0?void 0:p.y)!==null&&u!==void 0?u:0}})}}),n}return rU(t,e),JG(t,[{key:"updateBBox",value:function(){if(this.wrapperNode&&this.wrapperNode.getBoundingClientRect){var r=this.wrapperNode.getBoundingClientRect();(Math.abs(r.width-this.state.lastBoundingBox.width)>Dw||Math.abs(r.height-this.state.lastBoundingBox.height)>Dw)&&this.setState({lastBoundingBox:{width:r.width,height:r.height}})}else(this.state.lastBoundingBox.width!==-1||this.state.lastBoundingBox.height!==-1)&&this.setState({lastBoundingBox:{width:-1,height:-1}})}},{key:"componentDidMount",value:function(){document.addEventListener("keydown",this.handleKeyDown),this.updateBBox()}},{key:"componentWillUnmount",value:function(){document.removeEventListener("keydown",this.handleKeyDown)}},{key:"componentDidUpdate",value:function(){var r,a;this.props.active&&this.updateBBox(),this.state.dismissed&&(((r=this.props.coordinate)===null||r===void 0?void 0:r.x)!==this.state.dismissedAtCoordinate.x||((a=this.props.coordinate)===null||a===void 0?void 0:a.y)!==this.state.dismissedAtCoordinate.y)&&(this.state.dismissed=!1)}},{key:"render",value:function(){var r=this,a=this.props,o=a.active,i=a.allowEscapeViewBox,s=a.animationDuration,l=a.animationEasing,u=a.children,p=a.coordinate,c=a.hasPayload,f=a.isAnimationActive,m=a.offset,h=a.position,d=a.reverseDirection,g=a.useTranslate3d,v=a.viewBox,y=a.wrapperStyle,x=YG({allowEscapeViewBox:i,coordinate:p,offsetTopLeft:m,position:h,reverseDirection:d,tooltipBox:this.state.lastBoundingBox,useTranslate3d:g,viewBox:v}),P=x.cssClasses,w=x.cssProperties,S=Iw(Iw({transition:f&&o?"transform ".concat(s,"ms ").concat(l):void 0},w),{},{pointerEvents:"none",visibility:!this.state.dismissed&&o&&c?"visible":"hidden",position:"absolute",top:0,left:0},y);return E.createElement("div",{tabIndex:-1,className:P,style:S,ref:function(C){r.wrapperNode=C}},u)}}])}(k.PureComponent),iU=function(){return!(typeof window<"u"&&window.document&&window.document.createElement&&window.setTimeout)},_o={isSsr:iU()};function Ri(e){"@babel/helpers - typeof";return Ri=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Ri(e)}function Lw(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Fw(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Lw(Object(n),!0).forEach(function(r){ng(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Lw(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function sU(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function lU(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,o7(r.key),r)}}function uU(e,t,n){return t&&lU(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function cU(e,t,n){return t=Ap(t),pU(e,a7()?Reflect.construct(t,n||[],Ap(e).constructor):t.apply(e,n))}function pU(e,t){if(t&&(Ri(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return fU(e)}function fU(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function a7(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(a7=function(){return!!e})()}function Ap(e){return Ap=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},Ap(e)}function dU(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&x0(e,t)}function x0(e,t){return x0=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},x0(e,t)}function ng(e,t,n){return t=o7(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function o7(e){var t=mU(e,"string");return Ri(t)=="symbol"?t:t+""}function mU(e,t){if(Ri(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Ri(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}function hU(e){return e.dataKey}function vU(e,t){return E.isValidElement(e)?E.cloneElement(e,t):typeof e=="function"?E.createElement(e,t):E.createElement(UG,t)}var Yn=function(e){function t(){return sU(this,t),cU(this,t,arguments)}return dU(t,e),uU(t,[{key:"render",value:function(){var r=this,a=this.props,o=a.active,i=a.allowEscapeViewBox,s=a.animationDuration,l=a.animationEasing,u=a.content,p=a.coordinate,c=a.filterNull,f=a.isAnimationActive,m=a.offset,h=a.payload,d=a.payloadUniqBy,g=a.position,v=a.reverseDirection,y=a.useTranslate3d,x=a.viewBox,P=a.wrapperStyle,w=h??[];c&&w.length&&(w=K8(h.filter(function(O){return O.value!=null&&(O.hide!==!0||r.props.includeHidden)}),d,hU));var S=w.length>0;return E.createElement(oU,{allowEscapeViewBox:i,animationDuration:s,animationEasing:l,isAnimationActive:f,active:o,coordinate:p,hasPayload:S,offset:m,position:g,reverseDirection:v,useTranslate3d:y,viewBox:x,wrapperStyle:P},vU(u,Fw(Fw({},this.props),{},{payload:w})))}}])}(k.PureComponent);ng(Yn,"displayName","Tooltip");ng(Yn,"defaultProps",{accessibilityLayer:!1,allowEscapeViewBox:{x:!1,y:!1},animationDuration:400,animationEasing:"ease",contentStyle:{},coordinate:{x:0,y:0},cursor:!0,cursorStyle:{},filterNull:!0,isAnimationActive:!_o.isSsr,itemStyle:{},labelStyle:{},offset:10,reverseDirection:{x:!1,y:!1},separator:" : ",trigger:"hover",useTranslate3d:!1,viewBox:{x:0,y:0,height:0,width:0},wrapperStyle:{}});var yU=vr,gU=function(){return yU.Date.now()},xU=gU,wU=/\s/;function bU(e){for(var t=e.length;t--&&wU.test(e.charAt(t)););return t}var PU=bU,SU=PU,OU=/^\s+/;function kU(e){return e&&e.slice(0,SU(e)+1).replace(OU,"")}var CU=kU,_U=CU,Bw=Na,AU=ps,zw=NaN,EU=/^[-+]0x[0-9a-f]+$/i,TU=/^0b[01]+$/i,jU=/^0o[0-7]+$/i,NU=parseInt;function $U(e){if(typeof e=="number")return e;if(AU(e))return zw;if(Bw(e)){var t=typeof e.valueOf=="function"?e.valueOf():e;e=Bw(t)?t+"":t}if(typeof e!="string")return e===0?e:+e;e=_U(e);var n=TU.test(e);return n||jU.test(e)?NU(e.slice(2),n?2:8):EU.test(e)?zw:+e}var i7=$U,MU=Na,Cm=xU,Hw=i7,RU="Expected a function",IU=Math.max,DU=Math.min;function LU(e,t,n){var r,a,o,i,s,l,u=0,p=!1,c=!1,f=!0;if(typeof e!="function")throw new TypeError(RU);t=Hw(t)||0,MU(n)&&(p=!!n.leading,c="maxWait"in n,o=c?IU(Hw(n.maxWait)||0,t):o,f="trailing"in n?!!n.trailing:f);function m(S){var O=r,C=a;return r=a=void 0,u=S,i=e.apply(C,O),i}function h(S){return u=S,s=setTimeout(v,t),p?m(S):i}function d(S){var O=S-l,C=S-u,_=t-O;return c?DU(_,o-C):_}function g(S){var O=S-l,C=S-u;return l===void 0||O>=t||O<0||c&&C>=o}function v(){var S=Cm();if(g(S))return y(S);s=setTimeout(v,d(S))}function y(S){return s=void 0,f&&r?m(S):(r=a=void 0,i)}function x(){s!==void 0&&clearTimeout(s),u=0,r=l=a=s=void 0}function P(){return s===void 0?i:y(Cm())}function w(){var S=Cm(),O=g(S);if(r=arguments,a=this,l=S,O){if(s===void 0)return h(l);if(c)return clearTimeout(s),s=setTimeout(v,t),m(l)}return s===void 0&&(s=setTimeout(v,t)),i}return w.cancel=x,w.flush=P,w}var FU=LU,BU=FU,zU=Na,HU="Expected a function";function GU(e,t,n){var r=!0,a=!0;if(typeof e!="function")throw new TypeError(HU);return zU(n)&&(r="leading"in n?!!n.leading:r,a="trailing"in n?!!n.trailing:a),BU(e,t,{leading:r,maxWait:t,trailing:a})}var UU=GU;const s7=_e(UU);function Rl(e){"@babel/helpers - typeof";return Rl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Rl(e)}function Gw(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function ic(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Gw(Object(n),!0).forEach(function(r){WU(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Gw(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function WU(e,t,n){return t=qU(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function qU(e){var t=VU(e,"string");return Rl(t)=="symbol"?t:t+""}function VU(e,t){if(Rl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Rl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function KU(e,t){return ZU(e)||QU(e,t)||YU(e,t)||XU()}function XU(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function YU(e,t){if(e){if(typeof e=="string")return Uw(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return Uw(e,t)}}function Uw(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function QU(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function ZU(e){if(Array.isArray(e))return e}var JU=k.forwardRef(function(e,t){var n=e.aspect,r=e.initialDimension,a=r===void 0?{width:-1,height:-1}:r,o=e.width,i=o===void 0?"100%":o,s=e.height,l=s===void 0?"100%":s,u=e.minWidth,p=u===void 0?0:u,c=e.minHeight,f=e.maxHeight,m=e.children,h=e.debounce,d=h===void 0?0:h,g=e.id,v=e.className,y=e.onResize,x=e.style,P=x===void 0?{}:x,w=k.useRef(null),S=k.useRef();S.current=y,k.useImperativeHandle(t,function(){return Object.defineProperty(w.current,"current",{get:function(){return console.warn("The usage of ref.current.current is deprecated and will no longer be supported."),w.current},configurable:!0})});var O=k.useState({containerWidth:a.width,containerHeight:a.height}),C=KU(O,2),_=C[0],T=C[1],A=k.useCallback(function($,M){T(function(I){var R=Math.round($),L=Math.round(M);return I.containerWidth===R&&I.containerHeight===L?I:{containerWidth:R,containerHeight:L}})},[]);k.useEffect(function(){var $=function(D){var H,W=D[0].contentRect,G=W.width,Z=W.height;A(G,Z),(H=S.current)===null||H===void 0||H.call(S,G,Z)};d>0&&($=s7($,d,{trailing:!0,leading:!1}));var M=new ResizeObserver($),I=w.current.getBoundingClientRect(),R=I.width,L=I.height;return A(R,L),M.observe(w.current),function(){M.disconnect()}},[A,d]);var j=k.useMemo(function(){var $=_.containerWidth,M=_.containerHeight;if($<0||M<0)return null;Tr(Xa(i)||Xa(l),`The width(%s) and height(%s) are both fixed numbers,
       maybe you don't need to use a ResponsiveContainer.`,i,l),Tr(!n||n>0,"The aspect(%s) must be greater than zero.",n);var I=Xa(i)?$:i,R=Xa(l)?M:l;n&&n>0&&(I?R=I/n:R&&(I=R*n),f&&R>f&&(R=f)),Tr(I>0||R>0,`The width(%s) and height(%s) of chart should be greater than 0,
       please check the style of container, or the props width(%s) and height(%s),
       or add a minWidth(%s) or minHeight(%s) or use aspect(%s) to control the
       height and width.`,I,R,i,l,p,c,n);var L=!Array.isArray(m)&&Er(m.type).endsWith("Chart");return E.Children.map(m,function(N){return E.isValidElement(N)?k.cloneElement(N,ic({width:I,height:R},L?{style:ic({height:"100%",width:"100%",maxHeight:R,maxWidth:I},N.props.style)}:{})):N})},[n,m,l,f,c,p,_,i]);return E.createElement("div",{id:g?"".concat(g):void 0,className:ue("recharts-responsive-container",v),style:ic(ic({},P),{},{width:i,height:l,minWidth:p,minHeight:c,maxHeight:f}),ref:w},j)}),l7=function(t){return null};l7.displayName="Cell";function Il(e){"@babel/helpers - typeof";return Il=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Il(e)}function Ww(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function w0(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Ww(Object(n),!0).forEach(function(r){eW(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Ww(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function eW(e,t,n){return t=tW(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function tW(e){var t=nW(e,"string");return Il(t)=="symbol"?t:t+""}function nW(e,t){if(Il(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Il(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var Ro={widthCache:{},cacheCount:0},rW=2e3,aW={position:"absolute",top:"-20000px",left:0,padding:0,margin:0,border:"none",whiteSpace:"pre"},qw="recharts_measurement_span";function oW(e){var t=w0({},e);return Object.keys(t).forEach(function(n){t[n]||delete t[n]}),t}var il=function(t){var n=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{};if(t==null||_o.isSsr)return{width:0,height:0};var r=oW(n),a=JSON.stringify({text:t,copyStyle:r});if(Ro.widthCache[a])return Ro.widthCache[a];try{var o=document.getElementById(qw);o||(o=document.createElement("span"),o.setAttribute("id",qw),o.setAttribute("aria-hidden","true"),document.body.appendChild(o));var i=w0(w0({},aW),r);Object.assign(o.style,i),o.textContent="".concat(t);var s=o.getBoundingClientRect(),l={width:s.width,height:s.height};return Ro.widthCache[a]=l,++Ro.cacheCount>rW&&(Ro.cacheCount=0,Ro.widthCache={}),l}catch{return{width:0,height:0}}},iW=function(t){return{top:t.top+window.scrollY-document.documentElement.clientTop,left:t.left+window.scrollX-document.documentElement.clientLeft}};function Dl(e){"@babel/helpers - typeof";return Dl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Dl(e)}function Ep(e,t){return cW(e)||uW(e,t)||lW(e,t)||sW()}function sW(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function lW(e,t){if(e){if(typeof e=="string")return Vw(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return Vw(e,t)}}function Vw(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function uW(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t===0){if(Object(n)!==n)return;l=!1}else for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function cW(e){if(Array.isArray(e))return e}function pW(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Kw(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,dW(r.key),r)}}function fW(e,t,n){return t&&Kw(e.prototype,t),n&&Kw(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function dW(e){var t=mW(e,"string");return Dl(t)=="symbol"?t:t+""}function mW(e,t){if(Dl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Dl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var Xw=/(-?\d+(?:\.\d+)?[a-zA-Z%]*)([*/])(-?\d+(?:\.\d+)?[a-zA-Z%]*)/,Yw=/(-?\d+(?:\.\d+)?[a-zA-Z%]*)([+-])(-?\d+(?:\.\d+)?[a-zA-Z%]*)/,hW=/^px|cm|vh|vw|em|rem|%|mm|in|pt|pc|ex|ch|vmin|vmax|Q$/,vW=/(-?\d+(?:\.\d+)?)([a-zA-Z%]+)?/,u7={cm:96/2.54,mm:96/25.4,pt:96/72,pc:96/6,in:96,Q:96/(2.54*40),px:1},yW=Object.keys(u7),Qo="NaN";function gW(e,t){return e*u7[t]}var sc=function(){function e(t,n){pW(this,e),this.num=t,this.unit=n,this.num=t,this.unit=n,Number.isNaN(t)&&(this.unit=""),n!==""&&!hW.test(n)&&(this.num=NaN,this.unit=""),yW.includes(n)&&(this.num=gW(t,n),this.unit="px")}return fW(e,[{key:"add",value:function(n){return this.unit!==n.unit?new e(NaN,""):new e(this.num+n.num,this.unit)}},{key:"subtract",value:function(n){return this.unit!==n.unit?new e(NaN,""):new e(this.num-n.num,this.unit)}},{key:"multiply",value:function(n){return this.unit!==""&&n.unit!==""&&this.unit!==n.unit?new e(NaN,""):new e(this.num*n.num,this.unit||n.unit)}},{key:"divide",value:function(n){return this.unit!==""&&n.unit!==""&&this.unit!==n.unit?new e(NaN,""):new e(this.num/n.num,this.unit||n.unit)}},{key:"toString",value:function(){return"".concat(this.num).concat(this.unit)}},{key:"isNaN",value:function(){return Number.isNaN(this.num)}}],[{key:"parse",value:function(n){var r,a=(r=vW.exec(n))!==null&&r!==void 0?r:[],o=Ep(a,3),i=o[1],s=o[2];return new e(parseFloat(i),s??"")}}])}();function c7(e){if(e.includes(Qo))return Qo;for(var t=e;t.includes("*")||t.includes("/");){var n,r=(n=Xw.exec(t))!==null&&n!==void 0?n:[],a=Ep(r,4),o=a[1],i=a[2],s=a[3],l=sc.parse(o??""),u=sc.parse(s??""),p=i==="*"?l.multiply(u):l.divide(u);if(p.isNaN())return Qo;t=t.replace(Xw,p.toString())}for(;t.includes("+")||/.-\d+(?:\.\d+)?/.test(t);){var c,f=(c=Yw.exec(t))!==null&&c!==void 0?c:[],m=Ep(f,4),h=m[1],d=m[2],g=m[3],v=sc.parse(h??""),y=sc.parse(g??""),x=d==="+"?v.add(y):v.subtract(y);if(x.isNaN())return Qo;t=t.replace(Yw,x.toString())}return t}var Qw=/\(([^()]*)\)/;function xW(e){for(var t=e;t.includes("(");){var n=Qw.exec(t),r=Ep(n,2),a=r[1];t=t.replace(Qw,c7(a))}return t}function wW(e){var t=e.replace(/\s+/g,"");return t=xW(t),t=c7(t),t}function bW(e){try{return wW(e)}catch{return Qo}}function _m(e){var t=bW(e.slice(5,-1));return t===Qo?"":t}var PW=["x","y","lineHeight","capHeight","scaleToFit","textAnchor","verticalAnchor","fill"],SW=["dx","dy","angle","className","breakAll"];function b0(){return b0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},b0.apply(this,arguments)}function Zw(e,t){if(e==null)return{};var n=OW(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function OW(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function Jw(e,t){return AW(e)||_W(e,t)||CW(e,t)||kW()}function kW(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function CW(e,t){if(e){if(typeof e=="string")return e3(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return e3(e,t)}}function e3(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function _W(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t===0){if(Object(n)!==n)return;l=!1}else for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function AW(e){if(Array.isArray(e))return e}var p7=/[ \f\n\r\t\v\u2028\u2029]+/,f7=function(t){var n=t.children,r=t.breakAll,a=t.style;try{var o=[];le(n)||(r?o=n.toString().split(""):o=n.toString().split(p7));var i=o.map(function(l){return{word:l,width:il(l,a).width}}),s=r?0:il(" ",a).width;return{wordsWithComputedWidth:i,spaceWidth:s}}catch{return null}},EW=function(t,n,r,a,o){var i=t.maxLines,s=t.children,l=t.style,u=t.breakAll,p=V(i),c=s,f=function(){var I=arguments.length>0&&arguments[0]!==void 0?arguments[0]:[];return I.reduce(function(R,L){var N=L.word,D=L.width,H=R[R.length-1];if(H&&(a==null||o||H.width+D+r<Number(a)))H.words.push(N),H.width+=D+r;else{var W={words:[N],width:D};R.push(W)}return R},[])},m=f(n),h=function(I){return I.reduce(function(R,L){return R.width>L.width?R:L})};if(!p)return m;for(var d="…",g=function(I){var R=c.slice(0,I),L=f7({breakAll:u,style:l,children:R+d}).wordsWithComputedWidth,N=f(L),D=N.length>i||h(N).width>Number(a);return[D,N]},v=0,y=c.length-1,x=0,P;v<=y&&x<=c.length-1;){var w=Math.floor((v+y)/2),S=w-1,O=g(S),C=Jw(O,2),_=C[0],T=C[1],A=g(w),j=Jw(A,1),$=j[0];if(!_&&!$&&(v=w+1),_&&$&&(y=w-1),!_&&$){P=T;break}x++}return P||m},t3=function(t){var n=le(t)?[]:t.toString().split(p7);return[{words:n}]},TW=function(t){var n=t.width,r=t.scaleToFit,a=t.children,o=t.style,i=t.breakAll,s=t.maxLines;if((n||r)&&!_o.isSsr){var l,u,p=f7({breakAll:i,children:a,style:o});if(p){var c=p.wordsWithComputedWidth,f=p.spaceWidth;l=c,u=f}else return t3(a);return EW({breakAll:i,children:a,maxLines:s,style:o},l,u,n,r)}return t3(a)},n3="#808080",Tp=function(t){var n=t.x,r=n===void 0?0:n,a=t.y,o=a===void 0?0:a,i=t.lineHeight,s=i===void 0?"1em":i,l=t.capHeight,u=l===void 0?"0.71em":l,p=t.scaleToFit,c=p===void 0?!1:p,f=t.textAnchor,m=f===void 0?"start":f,h=t.verticalAnchor,d=h===void 0?"end":h,g=t.fill,v=g===void 0?n3:g,y=Zw(t,PW),x=k.useMemo(function(){return TW({breakAll:y.breakAll,children:y.children,maxLines:y.maxLines,scaleToFit:c,style:y.style,width:y.width})},[y.breakAll,y.children,y.maxLines,c,y.style,y.width]),P=y.dx,w=y.dy,S=y.angle,O=y.className,C=y.breakAll,_=Zw(y,SW);if(!ot(r)||!ot(o))return null;var T=r+(V(P)?P:0),A=o+(V(w)?w:0),j;switch(d){case"start":j=_m("calc(".concat(u,")"));break;case"middle":j=_m("calc(".concat((x.length-1)/2," * -").concat(s," + (").concat(u," / 2))"));break;default:j=_m("calc(".concat(x.length-1," * -").concat(s,")"));break}var $=[];if(c){var M=x[0].width,I=y.width;$.push("scale(".concat((V(I)?I/M:1)/M,")"))}return S&&$.push("rotate(".concat(S,", ").concat(T,", ").concat(A,")")),$.length&&(_.transform=$.join(" ")),E.createElement("text",b0({},ie(_,!0),{x:T,y:A,className:ue("recharts-text",O),textAnchor:m,fill:v.includes("url")?n3:v}),x.map(function(R,L){var N=R.words.join(C?"":" ");return E.createElement("tspan",{x:T,dy:L===0?j:s,key:"".concat(N,"-").concat(L)},N)}))};function ba(e,t){return e==null||t==null?NaN:e<t?-1:e>t?1:e>=t?0:NaN}function jW(e,t){return e==null||t==null?NaN:t<e?-1:t>e?1:t>=e?0:NaN}function rg(e){let t,n,r;e.length!==2?(t=ba,n=(s,l)=>ba(e(s),l),r=(s,l)=>e(s)-l):(t=e===ba||e===jW?e:NW,n=e,r=e);function a(s,l,u=0,p=s.length){if(u<p){if(t(l,l)!==0)return p;do{const c=u+p>>>1;n(s[c],l)<0?u=c+1:p=c}while(u<p)}return u}function o(s,l,u=0,p=s.length){if(u<p){if(t(l,l)!==0)return p;do{const c=u+p>>>1;n(s[c],l)<=0?u=c+1:p=c}while(u<p)}return u}function i(s,l,u=0,p=s.length){const c=a(s,l,u,p-1);return c>u&&r(s[c-1],l)>-r(s[c],l)?c-1:c}return{left:a,center:i,right:o}}function NW(){return 0}function d7(e){return e===null?NaN:+e}function*$W(e,t){for(let n of e)n!=null&&(n=+n)>=n&&(yield n)}const MW=rg(ba),Cu=MW.right;rg(d7).center;class r3 extends Map{constructor(t,n=DW){if(super(),Object.defineProperties(this,{_intern:{value:new Map},_key:{value:n}}),t!=null)for(const[r,a]of t)this.set(r,a)}get(t){return super.get(a3(this,t))}has(t){return super.has(a3(this,t))}set(t,n){return super.set(RW(this,t),n)}delete(t){return super.delete(IW(this,t))}}function a3({_intern:e,_key:t},n){const r=t(n);return e.has(r)?e.get(r):n}function RW({_intern:e,_key:t},n){const r=t(n);return e.has(r)?e.get(r):(e.set(r,n),n)}function IW({_intern:e,_key:t},n){const r=t(n);return e.has(r)&&(n=e.get(r),e.delete(r)),n}function DW(e){return e!==null&&typeof e=="object"?e.valueOf():e}function LW(e=ba){if(e===ba)return m7;if(typeof e!="function")throw new TypeError("compare is not a function");return(t,n)=>{const r=e(t,n);return r||r===0?r:(e(n,n)===0)-(e(t,t)===0)}}function m7(e,t){return(e==null||!(e>=e))-(t==null||!(t>=t))||(e<t?-1:e>t?1:0)}const FW=Math.sqrt(50),BW=Math.sqrt(10),zW=Math.sqrt(2);function jp(e,t,n){const r=(t-e)/Math.max(0,n),a=Math.floor(Math.log10(r)),o=r/Math.pow(10,a),i=o>=FW?10:o>=BW?5:o>=zW?2:1;let s,l,u;return a<0?(u=Math.pow(10,-a)/i,s=Math.round(e*u),l=Math.round(t*u),s/u<e&&++s,l/u>t&&--l,u=-u):(u=Math.pow(10,a)*i,s=Math.round(e/u),l=Math.round(t/u),s*u<e&&++s,l*u>t&&--l),l<s&&.5<=n&&n<2?jp(e,t,n*2):[s,l,u]}function P0(e,t,n){if(t=+t,e=+e,n=+n,!(n>0))return[];if(e===t)return[e];const r=t<e,[a,o,i]=r?jp(t,e,n):jp(e,t,n);if(!(o>=a))return[];const s=o-a+1,l=new Array(s);if(r)if(i<0)for(let u=0;u<s;++u)l[u]=(o-u)/-i;else for(let u=0;u<s;++u)l[u]=(o-u)*i;else if(i<0)for(let u=0;u<s;++u)l[u]=(a+u)/-i;else for(let u=0;u<s;++u)l[u]=(a+u)*i;return l}function S0(e,t,n){return t=+t,e=+e,n=+n,jp(e,t,n)[2]}function O0(e,t,n){t=+t,e=+e,n=+n;const r=t<e,a=r?S0(t,e,n):S0(e,t,n);return(r?-1:1)*(a<0?1/-a:a)}function o3(e,t){let n;for(const r of e)r!=null&&(n<r||n===void 0&&r>=r)&&(n=r);return n}function i3(e,t){let n;for(const r of e)r!=null&&(n>r||n===void 0&&r>=r)&&(n=r);return n}function h7(e,t,n=0,r=1/0,a){if(t=Math.floor(t),n=Math.floor(Math.max(0,n)),r=Math.floor(Math.min(e.length-1,r)),!(n<=t&&t<=r))return e;for(a=a===void 0?m7:LW(a);r>n;){if(r-n>600){const l=r-n+1,u=t-n+1,p=Math.log(l),c=.5*Math.exp(2*p/3),f=.5*Math.sqrt(p*c*(l-c)/l)*(u-l/2<0?-1:1),m=Math.max(n,Math.floor(t-u*c/l+f)),h=Math.min(r,Math.floor(t+(l-u)*c/l+f));h7(e,t,m,h,a)}const o=e[t];let i=n,s=r;for(Ds(e,n,t),a(e[r],o)>0&&Ds(e,n,r);i<s;){for(Ds(e,i,s),++i,--s;a(e[i],o)<0;)++i;for(;a(e[s],o)>0;)--s}a(e[n],o)===0?Ds(e,n,s):(++s,Ds(e,s,r)),s<=t&&(n=s+1),t<=s&&(r=s-1)}return e}function Ds(e,t,n){const r=e[t];e[t]=e[n],e[n]=r}function HW(e,t,n){if(e=Float64Array.from($W(e)),!(!(r=e.length)||isNaN(t=+t))){if(t<=0||r<2)return i3(e);if(t>=1)return o3(e);var r,a=(r-1)*t,o=Math.floor(a),i=o3(h7(e,o).subarray(0,o+1)),s=i3(e.subarray(o+1));return i+(s-i)*(a-o)}}function GW(e,t,n=d7){if(!(!(r=e.length)||isNaN(t=+t))){if(t<=0||r<2)return+n(e[0],0,e);if(t>=1)return+n(e[r-1],r-1,e);var r,a=(r-1)*t,o=Math.floor(a),i=+n(e[o],o,e),s=+n(e[o+1],o+1,e);return i+(s-i)*(a-o)}}function UW(e,t,n){e=+e,t=+t,n=(a=arguments.length)<2?(t=e,e=0,1):a<3?1:+n;for(var r=-1,a=Math.max(0,Math.ceil((t-e)/n))|0,o=new Array(a);++r<a;)o[r]=e+r*n;return o}function Pn(e,t){switch(arguments.length){case 0:break;case 1:this.range(e);break;default:this.range(t).domain(e);break}return this}function Wr(e,t){switch(arguments.length){case 0:break;case 1:{typeof e=="function"?this.interpolator(e):this.range(e);break}default:{this.domain(e),typeof t=="function"?this.interpolator(t):this.range(t);break}}return this}const k0=Symbol("implicit");function ag(){var e=new r3,t=[],n=[],r=k0;function a(o){let i=e.get(o);if(i===void 0){if(r!==k0)return r;e.set(o,i=t.push(o)-1)}return n[i%n.length]}return a.domain=function(o){if(!arguments.length)return t.slice();t=[],e=new r3;for(const i of o)e.has(i)||e.set(i,t.push(i)-1);return a},a.range=function(o){return arguments.length?(n=Array.from(o),a):n.slice()},a.unknown=function(o){return arguments.length?(r=o,a):r},a.copy=function(){return ag(t,n).unknown(r)},Pn.apply(a,arguments),a}function Ll(){var e=ag().unknown(void 0),t=e.domain,n=e.range,r=0,a=1,o,i,s=!1,l=0,u=0,p=.5;delete e.unknown;function c(){var f=t().length,m=a<r,h=m?a:r,d=m?r:a;o=(d-h)/Math.max(1,f-l+u*2),s&&(o=Math.floor(o)),h+=(d-h-o*(f-l))*p,i=o*(1-l),s&&(h=Math.round(h),i=Math.round(i));var g=UW(f).map(function(v){return h+o*v});return n(m?g.reverse():g)}return e.domain=function(f){return arguments.length?(t(f),c()):t()},e.range=function(f){return arguments.length?([r,a]=f,r=+r,a=+a,c()):[r,a]},e.rangeRound=function(f){return[r,a]=f,r=+r,a=+a,s=!0,c()},e.bandwidth=function(){return i},e.step=function(){return o},e.round=function(f){return arguments.length?(s=!!f,c()):s},e.padding=function(f){return arguments.length?(l=Math.min(1,u=+f),c()):l},e.paddingInner=function(f){return arguments.length?(l=Math.min(1,f),c()):l},e.paddingOuter=function(f){return arguments.length?(u=+f,c()):u},e.align=function(f){return arguments.length?(p=Math.max(0,Math.min(1,f)),c()):p},e.copy=function(){return Ll(t(),[r,a]).round(s).paddingInner(l).paddingOuter(u).align(p)},Pn.apply(c(),arguments)}function v7(e){var t=e.copy;return e.padding=e.paddingOuter,delete e.paddingInner,delete e.paddingOuter,e.copy=function(){return v7(t())},e}function sl(){return v7(Ll.apply(null,arguments).paddingInner(1))}function og(e,t,n){e.prototype=t.prototype=n,n.constructor=e}function y7(e,t){var n=Object.create(e.prototype);for(var r in t)n[r]=t[r];return n}function _u(){}var Fl=.7,Np=1/Fl,pi="\\s*([+-]?\\d+)\\s*",Bl="\\s*([+-]?(?:\\d*\\.)?\\d+(?:[eE][+-]?\\d+)?)\\s*",cr="\\s*([+-]?(?:\\d*\\.)?\\d+(?:[eE][+-]?\\d+)?)%\\s*",WW=/^#([0-9a-f]{3,8})$/,qW=new RegExp(`^rgb\\(${pi},${pi},${pi}\\)$`),VW=new RegExp(`^rgb\\(${cr},${cr},${cr}\\)$`),KW=new RegExp(`^rgba\\(${pi},${pi},${pi},${Bl}\\)$`),XW=new RegExp(`^rgba\\(${cr},${cr},${cr},${Bl}\\)$`),YW=new RegExp(`^hsl\\(${Bl},${cr},${cr}\\)$`),QW=new RegExp(`^hsla\\(${Bl},${cr},${cr},${Bl}\\)$`),s3={aliceblue:15792383,antiquewhite:16444375,aqua:65535,aquamarine:8388564,azure:15794175,beige:16119260,bisque:16770244,black:0,blanchedalmond:16772045,blue:255,blueviolet:9055202,brown:10824234,burlywood:14596231,cadetblue:6266528,chartreuse:8388352,chocolate:13789470,coral:16744272,cornflowerblue:6591981,cornsilk:16775388,crimson:14423100,cyan:65535,darkblue:139,darkcyan:35723,darkgoldenrod:12092939,darkgray:11119017,darkgreen:25600,darkgrey:11119017,darkkhaki:12433259,darkmagenta:9109643,darkolivegreen:5597999,darkorange:16747520,darkorchid:10040012,darkred:9109504,darksalmon:15308410,darkseagreen:9419919,darkslateblue:4734347,darkslategray:3100495,darkslategrey:3100495,darkturquoise:52945,darkviolet:9699539,deeppink:16716947,deepskyblue:49151,dimgray:6908265,dimgrey:6908265,dodgerblue:2003199,firebrick:11674146,floralwhite:16775920,forestgreen:2263842,fuchsia:16711935,gainsboro:14474460,ghostwhite:16316671,gold:16766720,goldenrod:14329120,gray:8421504,green:32768,greenyellow:11403055,grey:8421504,honeydew:15794160,hotpink:16738740,indianred:13458524,indigo:4915330,ivory:16777200,khaki:15787660,lavender:15132410,lavenderblush:16773365,lawngreen:8190976,lemonchiffon:16775885,lightblue:11393254,lightcoral:15761536,lightcyan:14745599,lightgoldenrodyellow:16448210,lightgray:13882323,lightgreen:9498256,lightgrey:13882323,lightpink:16758465,lightsalmon:16752762,lightseagreen:2142890,lightskyblue:8900346,lightslategray:7833753,lightslategrey:7833753,lightsteelblue:11584734,lightyellow:16777184,lime:65280,limegreen:3329330,linen:16445670,magenta:16711935,maroon:8388608,mediumaquamarine:6737322,mediumblue:205,mediumorchid:12211667,mediumpurple:9662683,mediumseagreen:3978097,mediumslateblue:8087790,mediumspringgreen:64154,mediumturquoise:4772300,mediumvioletred:13047173,midnightblue:1644912,mintcream:16121850,mistyrose:16770273,moccasin:16770229,navajowhite:16768685,navy:128,oldlace:16643558,olive:8421376,olivedrab:7048739,orange:16753920,orangered:16729344,orchid:14315734,palegoldenrod:15657130,palegreen:10025880,paleturquoise:11529966,palevioletred:14381203,papayawhip:16773077,peachpuff:16767673,peru:13468991,pink:16761035,plum:14524637,powderblue:11591910,purple:8388736,rebeccapurple:6697881,red:16711680,rosybrown:12357519,royalblue:4286945,saddlebrown:9127187,salmon:16416882,sandybrown:16032864,seagreen:3050327,seashell:16774638,sienna:10506797,silver:12632256,skyblue:8900331,slateblue:6970061,slategray:7372944,slategrey:7372944,snow:16775930,springgreen:65407,steelblue:4620980,tan:13808780,teal:32896,thistle:14204888,tomato:16737095,turquoise:4251856,violet:15631086,wheat:16113331,white:16777215,whitesmoke:16119285,yellow:16776960,yellowgreen:10145074};og(_u,zl,{copy(e){return Object.assign(new this.constructor,this,e)},displayable(){return this.rgb().displayable()},hex:l3,formatHex:l3,formatHex8:ZW,formatHsl:JW,formatRgb:u3,toString:u3});function l3(){return this.rgb().formatHex()}function ZW(){return this.rgb().formatHex8()}function JW(){return g7(this).formatHsl()}function u3(){return this.rgb().formatRgb()}function zl(e){var t,n;return e=(e+"").trim().toLowerCase(),(t=WW.exec(e))?(n=t[1].length,t=parseInt(t[1],16),n===6?c3(t):n===3?new Ht(t>>8&15|t>>4&240,t>>4&15|t&240,(t&15)<<4|t&15,1):n===8?lc(t>>24&255,t>>16&255,t>>8&255,(t&255)/255):n===4?lc(t>>12&15|t>>8&240,t>>8&15|t>>4&240,t>>4&15|t&240,((t&15)<<4|t&15)/255):null):(t=qW.exec(e))?new Ht(t[1],t[2],t[3],1):(t=VW.exec(e))?new Ht(t[1]*255/100,t[2]*255/100,t[3]*255/100,1):(t=KW.exec(e))?lc(t[1],t[2],t[3],t[4]):(t=XW.exec(e))?lc(t[1]*255/100,t[2]*255/100,t[3]*255/100,t[4]):(t=YW.exec(e))?d3(t[1],t[2]/100,t[3]/100,1):(t=QW.exec(e))?d3(t[1],t[2]/100,t[3]/100,t[4]):s3.hasOwnProperty(e)?c3(s3[e]):e==="transparent"?new Ht(NaN,NaN,NaN,0):null}function c3(e){return new Ht(e>>16&255,e>>8&255,e&255,1)}function lc(e,t,n,r){return r<=0&&(e=t=n=NaN),new Ht(e,t,n,r)}function eq(e){return e instanceof _u||(e=zl(e)),e?(e=e.rgb(),new Ht(e.r,e.g,e.b,e.opacity)):new Ht}function C0(e,t,n,r){return arguments.length===1?eq(e):new Ht(e,t,n,r??1)}function Ht(e,t,n,r){this.r=+e,this.g=+t,this.b=+n,this.opacity=+r}og(Ht,C0,y7(_u,{brighter(e){return e=e==null?Np:Math.pow(Np,e),new Ht(this.r*e,this.g*e,this.b*e,this.opacity)},darker(e){return e=e==null?Fl:Math.pow(Fl,e),new Ht(this.r*e,this.g*e,this.b*e,this.opacity)},rgb(){return this},clamp(){return new Ht(lo(this.r),lo(this.g),lo(this.b),$p(this.opacity))},displayable(){return-.5<=this.r&&this.r<255.5&&-.5<=this.g&&this.g<255.5&&-.5<=this.b&&this.b<255.5&&0<=this.opacity&&this.opacity<=1},hex:p3,formatHex:p3,formatHex8:tq,formatRgb:f3,toString:f3}));function p3(){return`#${Ya(this.r)}${Ya(this.g)}${Ya(this.b)}`}function tq(){return`#${Ya(this.r)}${Ya(this.g)}${Ya(this.b)}${Ya((isNaN(this.opacity)?1:this.opacity)*255)}`}function f3(){const e=$p(this.opacity);return`${e===1?"rgb(":"rgba("}${lo(this.r)}, ${lo(this.g)}, ${lo(this.b)}${e===1?")":`, ${e})`}`}function $p(e){return isNaN(e)?1:Math.max(0,Math.min(1,e))}function lo(e){return Math.max(0,Math.min(255,Math.round(e)||0))}function Ya(e){return e=lo(e),(e<16?"0":"")+e.toString(16)}function d3(e,t,n,r){return r<=0?e=t=n=NaN:n<=0||n>=1?e=t=NaN:t<=0&&(e=NaN),new Rn(e,t,n,r)}function g7(e){if(e instanceof Rn)return new Rn(e.h,e.s,e.l,e.opacity);if(e instanceof _u||(e=zl(e)),!e)return new Rn;if(e instanceof Rn)return e;e=e.rgb();var t=e.r/255,n=e.g/255,r=e.b/255,a=Math.min(t,n,r),o=Math.max(t,n,r),i=NaN,s=o-a,l=(o+a)/2;return s?(t===o?i=(n-r)/s+(n<r)*6:n===o?i=(r-t)/s+2:i=(t-n)/s+4,s/=l<.5?o+a:2-o-a,i*=60):s=l>0&&l<1?0:i,new Rn(i,s,l,e.opacity)}function nq(e,t,n,r){return arguments.length===1?g7(e):new Rn(e,t,n,r??1)}function Rn(e,t,n,r){this.h=+e,this.s=+t,this.l=+n,this.opacity=+r}og(Rn,nq,y7(_u,{brighter(e){return e=e==null?Np:Math.pow(Np,e),new Rn(this.h,this.s,this.l*e,this.opacity)},darker(e){return e=e==null?Fl:Math.pow(Fl,e),new Rn(this.h,this.s,this.l*e,this.opacity)},rgb(){var e=this.h%360+(this.h<0)*360,t=isNaN(e)||isNaN(this.s)?0:this.s,n=this.l,r=n+(n<.5?n:1-n)*t,a=2*n-r;return new Ht(Am(e>=240?e-240:e+120,a,r),Am(e,a,r),Am(e<120?e+240:e-120,a,r),this.opacity)},clamp(){return new Rn(m3(this.h),uc(this.s),uc(this.l),$p(this.opacity))},displayable(){return(0<=this.s&&this.s<=1||isNaN(this.s))&&0<=this.l&&this.l<=1&&0<=this.opacity&&this.opacity<=1},formatHsl(){const e=$p(this.opacity);return`${e===1?"hsl(":"hsla("}${m3(this.h)}, ${uc(this.s)*100}%, ${uc(this.l)*100}%${e===1?")":`, ${e})`}`}}));function m3(e){return e=(e||0)%360,e<0?e+360:e}function uc(e){return Math.max(0,Math.min(1,e||0))}function Am(e,t,n){return(e<60?t+(n-t)*e/60:e<180?n:e<240?t+(n-t)*(240-e)/60:t)*255}const ig=e=>()=>e;function rq(e,t){return function(n){return e+n*t}}function aq(e,t,n){return e=Math.pow(e,n),t=Math.pow(t,n)-e,n=1/n,function(r){return Math.pow(e+r*t,n)}}function oq(e){return(e=+e)==1?x7:function(t,n){return n-t?aq(t,n,e):ig(isNaN(t)?n:t)}}function x7(e,t){var n=t-e;return n?rq(e,n):ig(isNaN(e)?t:e)}const h3=function e(t){var n=oq(t);function r(a,o){var i=n((a=C0(a)).r,(o=C0(o)).r),s=n(a.g,o.g),l=n(a.b,o.b),u=x7(a.opacity,o.opacity);return function(p){return a.r=i(p),a.g=s(p),a.b=l(p),a.opacity=u(p),a+""}}return r.gamma=e,r}(1);function iq(e,t){t||(t=[]);var n=e?Math.min(t.length,e.length):0,r=t.slice(),a;return function(o){for(a=0;a<n;++a)r[a]=e[a]*(1-o)+t[a]*o;return r}}function sq(e){return ArrayBuffer.isView(e)&&!(e instanceof DataView)}function lq(e,t){var n=t?t.length:0,r=e?Math.min(n,e.length):0,a=new Array(r),o=new Array(n),i;for(i=0;i<r;++i)a[i]=bs(e[i],t[i]);for(;i<n;++i)o[i]=t[i];return function(s){for(i=0;i<r;++i)o[i]=a[i](s);return o}}function uq(e,t){var n=new Date;return e=+e,t=+t,function(r){return n.setTime(e*(1-r)+t*r),n}}function Mp(e,t){return e=+e,t=+t,function(n){return e*(1-n)+t*n}}function cq(e,t){var n={},r={},a;(e===null||typeof e!="object")&&(e={}),(t===null||typeof t!="object")&&(t={});for(a in t)a in e?n[a]=bs(e[a],t[a]):r[a]=t[a];return function(o){for(a in n)r[a]=n[a](o);return r}}var _0=/[-+]?(?:\d+\.?\d*|\.?\d+)(?:[eE][-+]?\d+)?/g,Em=new RegExp(_0.source,"g");function pq(e){return function(){return e}}function fq(e){return function(t){return e(t)+""}}function dq(e,t){var n=_0.lastIndex=Em.lastIndex=0,r,a,o,i=-1,s=[],l=[];for(e=e+"",t=t+"";(r=_0.exec(e))&&(a=Em.exec(t));)(o=a.index)>n&&(o=t.slice(n,o),s[i]?s[i]+=o:s[++i]=o),(r=r[0])===(a=a[0])?s[i]?s[i]+=a:s[++i]=a:(s[++i]=null,l.push({i,x:Mp(r,a)})),n=Em.lastIndex;return n<t.length&&(o=t.slice(n),s[i]?s[i]+=o:s[++i]=o),s.length<2?l[0]?fq(l[0].x):pq(t):(t=l.length,function(u){for(var p=0,c;p<t;++p)s[(c=l[p]).i]=c.x(u);return s.join("")})}function bs(e,t){var n=typeof t,r;return t==null||n==="boolean"?ig(t):(n==="number"?Mp:n==="string"?(r=zl(t))?(t=r,h3):dq:t instanceof zl?h3:t instanceof Date?uq:sq(t)?iq:Array.isArray(t)?lq:typeof t.valueOf!="function"&&typeof t.toString!="function"||isNaN(t)?cq:Mp)(e,t)}function sg(e,t){return e=+e,t=+t,function(n){return Math.round(e*(1-n)+t*n)}}function mq(e,t){t===void 0&&(t=e,e=bs);for(var n=0,r=t.length-1,a=t[0],o=new Array(r<0?0:r);n<r;)o[n]=e(a,a=t[++n]);return function(i){var s=Math.max(0,Math.min(r-1,Math.floor(i*=r)));return o[s](i-s)}}function hq(e){return function(){return e}}function Rp(e){return+e}var v3=[0,1];function $t(e){return e}function A0(e,t){return(t-=e=+e)?function(n){return(n-e)/t}:hq(isNaN(t)?NaN:.5)}function vq(e,t){var n;return e>t&&(n=e,e=t,t=n),function(r){return Math.max(e,Math.min(t,r))}}function yq(e,t,n){var r=e[0],a=e[1],o=t[0],i=t[1];return a<r?(r=A0(a,r),o=n(i,o)):(r=A0(r,a),o=n(o,i)),function(s){return o(r(s))}}function gq(e,t,n){var r=Math.min(e.length,t.length)-1,a=new Array(r),o=new Array(r),i=-1;for(e[r]<e[0]&&(e=e.slice().reverse(),t=t.slice().reverse());++i<r;)a[i]=A0(e[i],e[i+1]),o[i]=n(t[i],t[i+1]);return function(s){var l=Cu(e,s,1,r)-1;return o[l](a[l](s))}}function Au(e,t){return t.domain(e.domain()).range(e.range()).interpolate(e.interpolate()).clamp(e.clamp()).unknown(e.unknown())}function ud(){var e=v3,t=v3,n=bs,r,a,o,i=$t,s,l,u;function p(){var f=Math.min(e.length,t.length);return i!==$t&&(i=vq(e[0],e[f-1])),s=f>2?gq:yq,l=u=null,c}function c(f){return f==null||isNaN(f=+f)?o:(l||(l=s(e.map(r),t,n)))(r(i(f)))}return c.invert=function(f){return i(a((u||(u=s(t,e.map(r),Mp)))(f)))},c.domain=function(f){return arguments.length?(e=Array.from(f,Rp),p()):e.slice()},c.range=function(f){return arguments.length?(t=Array.from(f),p()):t.slice()},c.rangeRound=function(f){return t=Array.from(f),n=sg,p()},c.clamp=function(f){return arguments.length?(i=f?!0:$t,p()):i!==$t},c.interpolate=function(f){return arguments.length?(n=f,p()):n},c.unknown=function(f){return arguments.length?(o=f,c):o},function(f,m){return r=f,a=m,p()}}function lg(){return ud()($t,$t)}function xq(e){return Math.abs(e=Math.round(e))>=1e21?e.toLocaleString("en").replace(/,/g,""):e.toString(10)}function Ip(e,t){if((n=(e=t?e.toExponential(t-1):e.toExponential()).indexOf("e"))<0)return null;var n,r=e.slice(0,n);return[r.length>1?r[0]+r.slice(2):r,+e.slice(n+1)]}function Ii(e){return e=Ip(Math.abs(e)),e?e[1]:NaN}function wq(e,t){return function(n,r){for(var a=n.length,o=[],i=0,s=e[0],l=0;a>0&&s>0&&(l+s+1>r&&(s=Math.max(1,r-l)),o.push(n.substring(a-=s,a+s)),!((l+=s+1)>r));)s=e[i=(i+1)%e.length];return o.reverse().join(t)}}function bq(e){return function(t){return t.replace(/[0-9]/g,function(n){return e[+n]})}}var Pq=/^(?:(.)?([<>=^]))?([+\-( ])?([$#])?(0)?(\d+)?(,)?(\.\d+)?(~)?([a-z%])?$/i;function Hl(e){if(!(t=Pq.exec(e)))throw new Error("invalid format: "+e);var t;return new ug({fill:t[1],align:t[2],sign:t[3],symbol:t[4],zero:t[5],width:t[6],comma:t[7],precision:t[8]&&t[8].slice(1),trim:t[9],type:t[10]})}Hl.prototype=ug.prototype;function ug(e){this.fill=e.fill===void 0?" ":e.fill+"",this.align=e.align===void 0?">":e.align+"",this.sign=e.sign===void 0?"-":e.sign+"",this.symbol=e.symbol===void 0?"":e.symbol+"",this.zero=!!e.zero,this.width=e.width===void 0?void 0:+e.width,this.comma=!!e.comma,this.precision=e.precision===void 0?void 0:+e.precision,this.trim=!!e.trim,this.type=e.type===void 0?"":e.type+""}ug.prototype.toString=function(){return this.fill+this.align+this.sign+this.symbol+(this.zero?"0":"")+(this.width===void 0?"":Math.max(1,this.width|0))+(this.comma?",":"")+(this.precision===void 0?"":"."+Math.max(0,this.precision|0))+(this.trim?"~":"")+this.type};function Sq(e){e:for(var t=e.length,n=1,r=-1,a;n<t;++n)switch(e[n]){case".":r=a=n;break;case"0":r===0&&(r=n),a=n;break;default:if(!+e[n])break e;r>0&&(r=0);break}return r>0?e.slice(0,r)+e.slice(a+1):e}var w7;function Oq(e,t){var n=Ip(e,t);if(!n)return e+"";var r=n[0],a=n[1],o=a-(w7=Math.max(-8,Math.min(8,Math.floor(a/3)))*3)+1,i=r.length;return o===i?r:o>i?r+new Array(o-i+1).join("0"):o>0?r.slice(0,o)+"."+r.slice(o):"0."+new Array(1-o).join("0")+Ip(e,Math.max(0,t+o-1))[0]}function y3(e,t){var n=Ip(e,t);if(!n)return e+"";var r=n[0],a=n[1];return a<0?"0."+new Array(-a).join("0")+r:r.length>a+1?r.slice(0,a+1)+"."+r.slice(a+1):r+new Array(a-r.length+2).join("0")}const g3={"%":(e,t)=>(e*100).toFixed(t),b:e=>Math.round(e).toString(2),c:e=>e+"",d:xq,e:(e,t)=>e.toExponential(t),f:(e,t)=>e.toFixed(t),g:(e,t)=>e.toPrecision(t),o:e=>Math.round(e).toString(8),p:(e,t)=>y3(e*100,t),r:y3,s:Oq,X:e=>Math.round(e).toString(16).toUpperCase(),x:e=>Math.round(e).toString(16)};function x3(e){return e}var w3=Array.prototype.map,b3=["y","z","a","f","p","n","µ","m","","k","M","G","T","P","E","Z","Y"];function kq(e){var t=e.grouping===void 0||e.thousands===void 0?x3:wq(w3.call(e.grouping,Number),e.thousands+""),n=e.currency===void 0?"":e.currency[0]+"",r=e.currency===void 0?"":e.currency[1]+"",a=e.decimal===void 0?".":e.decimal+"",o=e.numerals===void 0?x3:bq(w3.call(e.numerals,String)),i=e.percent===void 0?"%":e.percent+"",s=e.minus===void 0?"−":e.minus+"",l=e.nan===void 0?"NaN":e.nan+"";function u(c){c=Hl(c);var f=c.fill,m=c.align,h=c.sign,d=c.symbol,g=c.zero,v=c.width,y=c.comma,x=c.precision,P=c.trim,w=c.type;w==="n"?(y=!0,w="g"):g3[w]||(x===void 0&&(x=12),P=!0,w="g"),(g||f==="0"&&m==="=")&&(g=!0,f="0",m="=");var S=d==="$"?n:d==="#"&&/[boxX]/.test(w)?"0"+w.toLowerCase():"",O=d==="$"?r:/[%p]/.test(w)?i:"",C=g3[w],_=/[defgprs%]/.test(w);x=x===void 0?6:/[gprs]/.test(w)?Math.max(1,Math.min(21,x)):Math.max(0,Math.min(20,x));function T(A){var j=S,$=O,M,I,R;if(w==="c")$=C(A)+$,A="";else{A=+A;var L=A<0||1/A<0;if(A=isNaN(A)?l:C(Math.abs(A),x),P&&(A=Sq(A)),L&&+A==0&&h!=="+"&&(L=!1),j=(L?h==="("?h:s:h==="-"||h==="("?"":h)+j,$=(w==="s"?b3[8+w7/3]:"")+$+(L&&h==="("?")":""),_){for(M=-1,I=A.length;++M<I;)if(R=A.charCodeAt(M),48>R||R>57){$=(R===46?a+A.slice(M+1):A.slice(M))+$,A=A.slice(0,M);break}}}y&&!g&&(A=t(A,1/0));var N=j.length+A.length+$.length,D=N<v?new Array(v-N+1).join(f):"";switch(y&&g&&(A=t(D+A,D.length?v-$.length:1/0),D=""),m){case"<":A=j+A+$+D;break;case"=":A=j+D+A+$;break;case"^":A=D.slice(0,N=D.length>>1)+j+A+$+D.slice(N);break;default:A=D+j+A+$;break}return o(A)}return T.toString=function(){return c+""},T}function p(c,f){var m=u((c=Hl(c),c.type="f",c)),h=Math.max(-8,Math.min(8,Math.floor(Ii(f)/3)))*3,d=Math.pow(10,-h),g=b3[8+h/3];return function(v){return m(d*v)+g}}return{format:u,formatPrefix:p}}var cc,cg,b7;Cq({thousands:",",grouping:[3],currency:["$",""]});function Cq(e){return cc=kq(e),cg=cc.format,b7=cc.formatPrefix,cc}function _q(e){return Math.max(0,-Ii(Math.abs(e)))}function Aq(e,t){return Math.max(0,Math.max(-8,Math.min(8,Math.floor(Ii(t)/3)))*3-Ii(Math.abs(e)))}function Eq(e,t){return e=Math.abs(e),t=Math.abs(t)-e,Math.max(0,Ii(t)-Ii(e))+1}function P7(e,t,n,r){var a=O0(e,t,n),o;switch(r=Hl(r??",f"),r.type){case"s":{var i=Math.max(Math.abs(e),Math.abs(t));return r.precision==null&&!isNaN(o=Aq(a,i))&&(r.precision=o),b7(r,i)}case"":case"e":case"g":case"p":case"r":{r.precision==null&&!isNaN(o=Eq(a,Math.max(Math.abs(e),Math.abs(t))))&&(r.precision=o-(r.type==="e"));break}case"f":case"%":{r.precision==null&&!isNaN(o=_q(a))&&(r.precision=o-(r.type==="%")*2);break}}return cg(r)}function Ma(e){var t=e.domain;return e.ticks=function(n){var r=t();return P0(r[0],r[r.length-1],n??10)},e.tickFormat=function(n,r){var a=t();return P7(a[0],a[a.length-1],n??10,r)},e.nice=function(n){n==null&&(n=10);var r=t(),a=0,o=r.length-1,i=r[a],s=r[o],l,u,p=10;for(s<i&&(u=i,i=s,s=u,u=a,a=o,o=u);p-- >0;){if(u=S0(i,s,n),u===l)return r[a]=i,r[o]=s,t(r);if(u>0)i=Math.floor(i/u)*u,s=Math.ceil(s/u)*u;else if(u<0)i=Math.ceil(i*u)/u,s=Math.floor(s*u)/u;else break;l=u}return e},e}function Dp(){var e=lg();return e.copy=function(){return Au(e,Dp())},Pn.apply(e,arguments),Ma(e)}function S7(e){var t;function n(r){return r==null||isNaN(r=+r)?t:r}return n.invert=n,n.domain=n.range=function(r){return arguments.length?(e=Array.from(r,Rp),n):e.slice()},n.unknown=function(r){return arguments.length?(t=r,n):t},n.copy=function(){return S7(e).unknown(t)},e=arguments.length?Array.from(e,Rp):[0,1],Ma(n)}function O7(e,t){e=e.slice();var n=0,r=e.length-1,a=e[n],o=e[r],i;return o<a&&(i=n,n=r,r=i,i=a,a=o,o=i),e[n]=t.floor(a),e[r]=t.ceil(o),e}function P3(e){return Math.log(e)}function S3(e){return Math.exp(e)}function Tq(e){return-Math.log(-e)}function jq(e){return-Math.exp(-e)}function Nq(e){return isFinite(e)?+("1e"+e):e<0?0:e}function $q(e){return e===10?Nq:e===Math.E?Math.exp:t=>Math.pow(e,t)}function Mq(e){return e===Math.E?Math.log:e===10&&Math.log10||e===2&&Math.log2||(e=Math.log(e),t=>Math.log(t)/e)}function O3(e){return(t,n)=>-e(-t,n)}function pg(e){const t=e(P3,S3),n=t.domain;let r=10,a,o;function i(){return a=Mq(r),o=$q(r),n()[0]<0?(a=O3(a),o=O3(o),e(Tq,jq)):e(P3,S3),t}return t.base=function(s){return arguments.length?(r=+s,i()):r},t.domain=function(s){return arguments.length?(n(s),i()):n()},t.ticks=s=>{const l=n();let u=l[0],p=l[l.length-1];const c=p<u;c&&([u,p]=[p,u]);let f=a(u),m=a(p),h,d;const g=s==null?10:+s;let v=[];if(!(r%1)&&m-f<g){if(f=Math.floor(f),m=Math.ceil(m),u>0){for(;f<=m;++f)for(h=1;h<r;++h)if(d=f<0?h/o(-f):h*o(f),!(d<u)){if(d>p)break;v.push(d)}}else for(;f<=m;++f)for(h=r-1;h>=1;--h)if(d=f>0?h/o(-f):h*o(f),!(d<u)){if(d>p)break;v.push(d)}v.length*2<g&&(v=P0(u,p,g))}else v=P0(f,m,Math.min(m-f,g)).map(o);return c?v.reverse():v},t.tickFormat=(s,l)=>{if(s==null&&(s=10),l==null&&(l=r===10?"s":","),typeof l!="function"&&(!(r%1)&&(l=Hl(l)).precision==null&&(l.trim=!0),l=cg(l)),s===1/0)return l;const u=Math.max(1,r*s/t.ticks().length);return p=>{let c=p/o(Math.round(a(p)));return c*r<r-.5&&(c*=r),c<=u?l(p):""}},t.nice=()=>n(O7(n(),{floor:s=>o(Math.floor(a(s))),ceil:s=>o(Math.ceil(a(s)))})),t}function k7(){const e=pg(ud()).domain([1,10]);return e.copy=()=>Au(e,k7()).base(e.base()),Pn.apply(e,arguments),e}function k3(e){return function(t){return Math.sign(t)*Math.log1p(Math.abs(t/e))}}function C3(e){return function(t){return Math.sign(t)*Math.expm1(Math.abs(t))*e}}function fg(e){var t=1,n=e(k3(t),C3(t));return n.constant=function(r){return arguments.length?e(k3(t=+r),C3(t)):t},Ma(n)}function C7(){var e=fg(ud());return e.copy=function(){return Au(e,C7()).constant(e.constant())},Pn.apply(e,arguments)}function _3(e){return function(t){return t<0?-Math.pow(-t,e):Math.pow(t,e)}}function Rq(e){return e<0?-Math.sqrt(-e):Math.sqrt(e)}function Iq(e){return e<0?-e*e:e*e}function dg(e){var t=e($t,$t),n=1;function r(){return n===1?e($t,$t):n===.5?e(Rq,Iq):e(_3(n),_3(1/n))}return t.exponent=function(a){return arguments.length?(n=+a,r()):n},Ma(t)}function mg(){var e=dg(ud());return e.copy=function(){return Au(e,mg()).exponent(e.exponent())},Pn.apply(e,arguments),e}function Dq(){return mg.apply(null,arguments).exponent(.5)}function A3(e){return Math.sign(e)*e*e}function Lq(e){return Math.sign(e)*Math.sqrt(Math.abs(e))}function _7(){var e=lg(),t=[0,1],n=!1,r;function a(o){var i=Lq(e(o));return isNaN(i)?r:n?Math.round(i):i}return a.invert=function(o){return e.invert(A3(o))},a.domain=function(o){return arguments.length?(e.domain(o),a):e.domain()},a.range=function(o){return arguments.length?(e.range((t=Array.from(o,Rp)).map(A3)),a):t.slice()},a.rangeRound=function(o){return a.range(o).round(!0)},a.round=function(o){return arguments.length?(n=!!o,a):n},a.clamp=function(o){return arguments.length?(e.clamp(o),a):e.clamp()},a.unknown=function(o){return arguments.length?(r=o,a):r},a.copy=function(){return _7(e.domain(),t).round(n).clamp(e.clamp()).unknown(r)},Pn.apply(a,arguments),Ma(a)}function A7(){var e=[],t=[],n=[],r;function a(){var i=0,s=Math.max(1,t.length);for(n=new Array(s-1);++i<s;)n[i-1]=GW(e,i/s);return o}function o(i){return i==null||isNaN(i=+i)?r:t[Cu(n,i)]}return o.invertExtent=function(i){var s=t.indexOf(i);return s<0?[NaN,NaN]:[s>0?n[s-1]:e[0],s<n.length?n[s]:e[e.length-1]]},o.domain=function(i){if(!arguments.length)return e.slice();e=[];for(let s of i)s!=null&&!isNaN(s=+s)&&e.push(s);return e.sort(ba),a()},o.range=function(i){return arguments.length?(t=Array.from(i),a()):t.slice()},o.unknown=function(i){return arguments.length?(r=i,o):r},o.quantiles=function(){return n.slice()},o.copy=function(){return A7().domain(e).range(t).unknown(r)},Pn.apply(o,arguments)}function E7(){var e=0,t=1,n=1,r=[.5],a=[0,1],o;function i(l){return l!=null&&l<=l?a[Cu(r,l,0,n)]:o}function s(){var l=-1;for(r=new Array(n);++l<n;)r[l]=((l+1)*t-(l-n)*e)/(n+1);return i}return i.domain=function(l){return arguments.length?([e,t]=l,e=+e,t=+t,s()):[e,t]},i.range=function(l){return arguments.length?(n=(a=Array.from(l)).length-1,s()):a.slice()},i.invertExtent=function(l){var u=a.indexOf(l);return u<0?[NaN,NaN]:u<1?[e,r[0]]:u>=n?[r[n-1],t]:[r[u-1],r[u]]},i.unknown=function(l){return arguments.length&&(o=l),i},i.thresholds=function(){return r.slice()},i.copy=function(){return E7().domain([e,t]).range(a).unknown(o)},Pn.apply(Ma(i),arguments)}function T7(){var e=[.5],t=[0,1],n,r=1;function a(o){return o!=null&&o<=o?t[Cu(e,o,0,r)]:n}return a.domain=function(o){return arguments.length?(e=Array.from(o),r=Math.min(e.length,t.length-1),a):e.slice()},a.range=function(o){return arguments.length?(t=Array.from(o),r=Math.min(e.length,t.length-1),a):t.slice()},a.invertExtent=function(o){var i=t.indexOf(o);return[e[i-1],e[i]]},a.unknown=function(o){return arguments.length?(n=o,a):n},a.copy=function(){return T7().domain(e).range(t).unknown(n)},Pn.apply(a,arguments)}const Tm=new Date,jm=new Date;function it(e,t,n,r){function a(o){return e(o=arguments.length===0?new Date:new Date(+o)),o}return a.floor=o=>(e(o=new Date(+o)),o),a.ceil=o=>(e(o=new Date(o-1)),t(o,1),e(o),o),a.round=o=>{const i=a(o),s=a.ceil(o);return o-i<s-o?i:s},a.offset=(o,i)=>(t(o=new Date(+o),i==null?1:Math.floor(i)),o),a.range=(o,i,s)=>{const l=[];if(o=a.ceil(o),s=s==null?1:Math.floor(s),!(o<i)||!(s>0))return l;let u;do l.push(u=new Date(+o)),t(o,s),e(o);while(u<o&&o<i);return l},a.filter=o=>it(i=>{if(i>=i)for(;e(i),!o(i);)i.setTime(i-1)},(i,s)=>{if(i>=i)if(s<0)for(;++s<=0;)for(;t(i,-1),!o(i););else for(;--s>=0;)for(;t(i,1),!o(i););}),n&&(a.count=(o,i)=>(Tm.setTime(+o),jm.setTime(+i),e(Tm),e(jm),Math.floor(n(Tm,jm))),a.every=o=>(o=Math.floor(o),!isFinite(o)||!(o>0)?null:o>1?a.filter(r?i=>r(i)%o===0:i=>a.count(0,i)%o===0):a)),a}const Lp=it(()=>{},(e,t)=>{e.setTime(+e+t)},(e,t)=>t-e);Lp.every=e=>(e=Math.floor(e),!isFinite(e)||!(e>0)?null:e>1?it(t=>{t.setTime(Math.floor(t/e)*e)},(t,n)=>{t.setTime(+t+n*e)},(t,n)=>(n-t)/e):Lp);Lp.range;const kr=1e3,mn=kr*60,Cr=mn*60,Fr=Cr*24,hg=Fr*7,E3=Fr*30,Nm=Fr*365,Qa=it(e=>{e.setTime(e-e.getMilliseconds())},(e,t)=>{e.setTime(+e+t*kr)},(e,t)=>(t-e)/kr,e=>e.getUTCSeconds());Qa.range;const vg=it(e=>{e.setTime(e-e.getMilliseconds()-e.getSeconds()*kr)},(e,t)=>{e.setTime(+e+t*mn)},(e,t)=>(t-e)/mn,e=>e.getMinutes());vg.range;const yg=it(e=>{e.setUTCSeconds(0,0)},(e,t)=>{e.setTime(+e+t*mn)},(e,t)=>(t-e)/mn,e=>e.getUTCMinutes());yg.range;const gg=it(e=>{e.setTime(e-e.getMilliseconds()-e.getSeconds()*kr-e.getMinutes()*mn)},(e,t)=>{e.setTime(+e+t*Cr)},(e,t)=>(t-e)/Cr,e=>e.getHours());gg.range;const xg=it(e=>{e.setUTCMinutes(0,0,0)},(e,t)=>{e.setTime(+e+t*Cr)},(e,t)=>(t-e)/Cr,e=>e.getUTCHours());xg.range;const Eu=it(e=>e.setHours(0,0,0,0),(e,t)=>e.setDate(e.getDate()+t),(e,t)=>(t-e-(t.getTimezoneOffset()-e.getTimezoneOffset())*mn)/Fr,e=>e.getDate()-1);Eu.range;const cd=it(e=>{e.setUTCHours(0,0,0,0)},(e,t)=>{e.setUTCDate(e.getUTCDate()+t)},(e,t)=>(t-e)/Fr,e=>e.getUTCDate()-1);cd.range;const j7=it(e=>{e.setUTCHours(0,0,0,0)},(e,t)=>{e.setUTCDate(e.getUTCDate()+t)},(e,t)=>(t-e)/Fr,e=>Math.floor(e/Fr));j7.range;function Ao(e){return it(t=>{t.setDate(t.getDate()-(t.getDay()+7-e)%7),t.setHours(0,0,0,0)},(t,n)=>{t.setDate(t.getDate()+n*7)},(t,n)=>(n-t-(n.getTimezoneOffset()-t.getTimezoneOffset())*mn)/hg)}const pd=Ao(0),Fp=Ao(1),Fq=Ao(2),Bq=Ao(3),Di=Ao(4),zq=Ao(5),Hq=Ao(6);pd.range;Fp.range;Fq.range;Bq.range;Di.range;zq.range;Hq.range;function Eo(e){return it(t=>{t.setUTCDate(t.getUTCDate()-(t.getUTCDay()+7-e)%7),t.setUTCHours(0,0,0,0)},(t,n)=>{t.setUTCDate(t.getUTCDate()+n*7)},(t,n)=>(n-t)/hg)}const fd=Eo(0),Bp=Eo(1),Gq=Eo(2),Uq=Eo(3),Li=Eo(4),Wq=Eo(5),qq=Eo(6);fd.range;Bp.range;Gq.range;Uq.range;Li.range;Wq.range;qq.range;const wg=it(e=>{e.setDate(1),e.setHours(0,0,0,0)},(e,t)=>{e.setMonth(e.getMonth()+t)},(e,t)=>t.getMonth()-e.getMonth()+(t.getFullYear()-e.getFullYear())*12,e=>e.getMonth());wg.range;const bg=it(e=>{e.setUTCDate(1),e.setUTCHours(0,0,0,0)},(e,t)=>{e.setUTCMonth(e.getUTCMonth()+t)},(e,t)=>t.getUTCMonth()-e.getUTCMonth()+(t.getUTCFullYear()-e.getUTCFullYear())*12,e=>e.getUTCMonth());bg.range;const Br=it(e=>{e.setMonth(0,1),e.setHours(0,0,0,0)},(e,t)=>{e.setFullYear(e.getFullYear()+t)},(e,t)=>t.getFullYear()-e.getFullYear(),e=>e.getFullYear());Br.every=e=>!isFinite(e=Math.floor(e))||!(e>0)?null:it(t=>{t.setFullYear(Math.floor(t.getFullYear()/e)*e),t.setMonth(0,1),t.setHours(0,0,0,0)},(t,n)=>{t.setFullYear(t.getFullYear()+n*e)});Br.range;const zr=it(e=>{e.setUTCMonth(0,1),e.setUTCHours(0,0,0,0)},(e,t)=>{e.setUTCFullYear(e.getUTCFullYear()+t)},(e,t)=>t.getUTCFullYear()-e.getUTCFullYear(),e=>e.getUTCFullYear());zr.every=e=>!isFinite(e=Math.floor(e))||!(e>0)?null:it(t=>{t.setUTCFullYear(Math.floor(t.getUTCFullYear()/e)*e),t.setUTCMonth(0,1),t.setUTCHours(0,0,0,0)},(t,n)=>{t.setUTCFullYear(t.getUTCFullYear()+n*e)});zr.range;function N7(e,t,n,r,a,o){const i=[[Qa,1,kr],[Qa,5,5*kr],[Qa,15,15*kr],[Qa,30,30*kr],[o,1,mn],[o,5,5*mn],[o,15,15*mn],[o,30,30*mn],[a,1,Cr],[a,3,3*Cr],[a,6,6*Cr],[a,12,12*Cr],[r,1,Fr],[r,2,2*Fr],[n,1,hg],[t,1,E3],[t,3,3*E3],[e,1,Nm]];function s(u,p,c){const f=p<u;f&&([u,p]=[p,u]);const m=c&&typeof c.range=="function"?c:l(u,p,c),h=m?m.range(u,+p+1):[];return f?h.reverse():h}function l(u,p,c){const f=Math.abs(p-u)/c,m=rg(([,,g])=>g).right(i,f);if(m===i.length)return e.every(O0(u/Nm,p/Nm,c));if(m===0)return Lp.every(Math.max(O0(u,p,c),1));const[h,d]=i[f/i[m-1][2]<i[m][2]/f?m-1:m];return h.every(d)}return[s,l]}const[Vq,Kq]=N7(zr,bg,fd,j7,xg,yg),[Xq,Yq]=N7(Br,wg,pd,Eu,gg,vg);function $m(e){if(0<=e.y&&e.y<100){var t=new Date(-1,e.m,e.d,e.H,e.M,e.S,e.L);return t.setFullYear(e.y),t}return new Date(e.y,e.m,e.d,e.H,e.M,e.S,e.L)}function Mm(e){if(0<=e.y&&e.y<100){var t=new Date(Date.UTC(-1,e.m,e.d,e.H,e.M,e.S,e.L));return t.setUTCFullYear(e.y),t}return new Date(Date.UTC(e.y,e.m,e.d,e.H,e.M,e.S,e.L))}function Ls(e,t,n){return{y:e,m:t,d:n,H:0,M:0,S:0,L:0}}function Qq(e){var t=e.dateTime,n=e.date,r=e.time,a=e.periods,o=e.days,i=e.shortDays,s=e.months,l=e.shortMonths,u=Fs(a),p=Bs(a),c=Fs(o),f=Bs(o),m=Fs(i),h=Bs(i),d=Fs(s),g=Bs(s),v=Fs(l),y=Bs(l),x={a:L,A:N,b:D,B:H,c:null,d:R3,e:R3,f:wV,g:TV,G:NV,H:yV,I:gV,j:xV,L:$7,m:bV,M:PV,p:W,q:G,Q:L3,s:F3,S:SV,u:OV,U:kV,V:CV,w:_V,W:AV,x:null,X:null,y:EV,Y:jV,Z:$V,"%":D3},P={a:Z,A:re,b:ve,B:be,c:null,d:I3,e:I3,f:DV,g:VV,G:XV,H:MV,I:RV,j:IV,L:R7,m:LV,M:FV,p:J,q:se,Q:L3,s:F3,S:BV,u:zV,U:HV,V:GV,w:UV,W:WV,x:null,X:null,y:qV,Y:KV,Z:YV,"%":D3},w={a:T,A,b:j,B:$,c:M,d:$3,e:$3,f:dV,g:N3,G:j3,H:M3,I:M3,j:uV,L:fV,m:lV,M:cV,p:_,q:sV,Q:hV,s:vV,S:pV,u:nV,U:rV,V:aV,w:tV,W:oV,x:I,X:R,y:N3,Y:j3,Z:iV,"%":mV};x.x=S(n,x),x.X=S(r,x),x.c=S(t,x),P.x=S(n,P),P.X=S(r,P),P.c=S(t,P);function S(q,K){return function(X){var F=[],pe=-1,te=0,$e=q.length,Me,Qe,Vn;for(X instanceof Date||(X=new Date(+X));++pe<$e;)q.charCodeAt(pe)===37&&(F.push(q.slice(te,pe)),(Qe=T3[Me=q.charAt(++pe)])!=null?Me=q.charAt(++pe):Qe=Me==="e"?" ":"0",(Vn=K[Me])&&(Me=Vn(X,Qe)),F.push(Me),te=pe+1);return F.push(q.slice(te,pe)),F.join("")}}function O(q,K){return function(X){var F=Ls(1900,void 0,1),pe=C(F,q,X+="",0),te,$e;if(pe!=X.length)return null;if("Q"in F)return new Date(F.Q);if("s"in F)return new Date(F.s*1e3+("L"in F?F.L:0));if(K&&!("Z"in F)&&(F.Z=0),"p"in F&&(F.H=F.H%12+F.p*12),F.m===void 0&&(F.m="q"in F?F.q:0),"V"in F){if(F.V<1||F.V>53)return null;"w"in F||(F.w=1),"Z"in F?(te=Mm(Ls(F.y,0,1)),$e=te.getUTCDay(),te=$e>4||$e===0?Bp.ceil(te):Bp(te),te=cd.offset(te,(F.V-1)*7),F.y=te.getUTCFullYear(),F.m=te.getUTCMonth(),F.d=te.getUTCDate()+(F.w+6)%7):(te=$m(Ls(F.y,0,1)),$e=te.getDay(),te=$e>4||$e===0?Fp.ceil(te):Fp(te),te=Eu.offset(te,(F.V-1)*7),F.y=te.getFullYear(),F.m=te.getMonth(),F.d=te.getDate()+(F.w+6)%7)}else("W"in F||"U"in F)&&("w"in F||(F.w="u"in F?F.u%7:"W"in F?1:0),$e="Z"in F?Mm(Ls(F.y,0,1)).getUTCDay():$m(Ls(F.y,0,1)).getDay(),F.m=0,F.d="W"in F?(F.w+6)%7+F.W*7-($e+5)%7:F.w+F.U*7-($e+6)%7);return"Z"in F?(F.H+=F.Z/100|0,F.M+=F.Z%100,Mm(F)):$m(F)}}function C(q,K,X,F){for(var pe=0,te=K.length,$e=X.length,Me,Qe;pe<te;){if(F>=$e)return-1;if(Me=K.charCodeAt(pe++),Me===37){if(Me=K.charAt(pe++),Qe=w[Me in T3?K.charAt(pe++):Me],!Qe||(F=Qe(q,X,F))<0)return-1}else if(Me!=X.charCodeAt(F++))return-1}return F}function _(q,K,X){var F=u.exec(K.slice(X));return F?(q.p=p.get(F[0].toLowerCase()),X+F[0].length):-1}function T(q,K,X){var F=m.exec(K.slice(X));return F?(q.w=h.get(F[0].toLowerCase()),X+F[0].length):-1}function A(q,K,X){var F=c.exec(K.slice(X));return F?(q.w=f.get(F[0].toLowerCase()),X+F[0].length):-1}function j(q,K,X){var F=v.exec(K.slice(X));return F?(q.m=y.get(F[0].toLowerCase()),X+F[0].length):-1}function $(q,K,X){var F=d.exec(K.slice(X));return F?(q.m=g.get(F[0].toLowerCase()),X+F[0].length):-1}function M(q,K,X){return C(q,t,K,X)}function I(q,K,X){return C(q,n,K,X)}function R(q,K,X){return C(q,r,K,X)}function L(q){return i[q.getDay()]}function N(q){return o[q.getDay()]}function D(q){return l[q.getMonth()]}function H(q){return s[q.getMonth()]}function W(q){return a[+(q.getHours()>=12)]}function G(q){return 1+~~(q.getMonth()/3)}function Z(q){return i[q.getUTCDay()]}function re(q){return o[q.getUTCDay()]}function ve(q){return l[q.getUTCMonth()]}function be(q){return s[q.getUTCMonth()]}function J(q){return a[+(q.getUTCHours()>=12)]}function se(q){return 1+~~(q.getUTCMonth()/3)}return{format:function(q){var K=S(q+="",x);return K.toString=function(){return q},K},parse:function(q){var K=O(q+="",!1);return K.toString=function(){return q},K},utcFormat:function(q){var K=S(q+="",P);return K.toString=function(){return q},K},utcParse:function(q){var K=O(q+="",!0);return K.toString=function(){return q},K}}}var T3={"-":"",_:" ",0:"0"},ct=/^\s*\d+/,Zq=/^%/,Jq=/[\\^$*+?|[\]().{}]/g;function he(e,t,n){var r=e<0?"-":"",a=(r?-e:e)+"",o=a.length;return r+(o<n?new Array(n-o+1).join(t)+a:a)}function eV(e){return e.replace(Jq,"\\$&")}function Fs(e){return new RegExp("^(?:"+e.map(eV).join("|")+")","i")}function Bs(e){return new Map(e.map((t,n)=>[t.toLowerCase(),n]))}function tV(e,t,n){var r=ct.exec(t.slice(n,n+1));return r?(e.w=+r[0],n+r[0].length):-1}function nV(e,t,n){var r=ct.exec(t.slice(n,n+1));return r?(e.u=+r[0],n+r[0].length):-1}function rV(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.U=+r[0],n+r[0].length):-1}function aV(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.V=+r[0],n+r[0].length):-1}function oV(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.W=+r[0],n+r[0].length):-1}function j3(e,t,n){var r=ct.exec(t.slice(n,n+4));return r?(e.y=+r[0],n+r[0].length):-1}function N3(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.y=+r[0]+(+r[0]>68?1900:2e3),n+r[0].length):-1}function iV(e,t,n){var r=/^(Z)|([+-]\d\d)(?::?(\d\d))?/.exec(t.slice(n,n+6));return r?(e.Z=r[1]?0:-(r[2]+(r[3]||"00")),n+r[0].length):-1}function sV(e,t,n){var r=ct.exec(t.slice(n,n+1));return r?(e.q=r[0]*3-3,n+r[0].length):-1}function lV(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.m=r[0]-1,n+r[0].length):-1}function $3(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.d=+r[0],n+r[0].length):-1}function uV(e,t,n){var r=ct.exec(t.slice(n,n+3));return r?(e.m=0,e.d=+r[0],n+r[0].length):-1}function M3(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.H=+r[0],n+r[0].length):-1}function cV(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.M=+r[0],n+r[0].length):-1}function pV(e,t,n){var r=ct.exec(t.slice(n,n+2));return r?(e.S=+r[0],n+r[0].length):-1}function fV(e,t,n){var r=ct.exec(t.slice(n,n+3));return r?(e.L=+r[0],n+r[0].length):-1}function dV(e,t,n){var r=ct.exec(t.slice(n,n+6));return r?(e.L=Math.floor(r[0]/1e3),n+r[0].length):-1}function mV(e,t,n){var r=Zq.exec(t.slice(n,n+1));return r?n+r[0].length:-1}function hV(e,t,n){var r=ct.exec(t.slice(n));return r?(e.Q=+r[0],n+r[0].length):-1}function vV(e,t,n){var r=ct.exec(t.slice(n));return r?(e.s=+r[0],n+r[0].length):-1}function R3(e,t){return he(e.getDate(),t,2)}function yV(e,t){return he(e.getHours(),t,2)}function gV(e,t){return he(e.getHours()%12||12,t,2)}function xV(e,t){return he(1+Eu.count(Br(e),e),t,3)}function $7(e,t){return he(e.getMilliseconds(),t,3)}function wV(e,t){return $7(e,t)+"000"}function bV(e,t){return he(e.getMonth()+1,t,2)}function PV(e,t){return he(e.getMinutes(),t,2)}function SV(e,t){return he(e.getSeconds(),t,2)}function OV(e){var t=e.getDay();return t===0?7:t}function kV(e,t){return he(pd.count(Br(e)-1,e),t,2)}function M7(e){var t=e.getDay();return t>=4||t===0?Di(e):Di.ceil(e)}function CV(e,t){return e=M7(e),he(Di.count(Br(e),e)+(Br(e).getDay()===4),t,2)}function _V(e){return e.getDay()}function AV(e,t){return he(Fp.count(Br(e)-1,e),t,2)}function EV(e,t){return he(e.getFullYear()%100,t,2)}function TV(e,t){return e=M7(e),he(e.getFullYear()%100,t,2)}function jV(e,t){return he(e.getFullYear()%1e4,t,4)}function NV(e,t){var n=e.getDay();return e=n>=4||n===0?Di(e):Di.ceil(e),he(e.getFullYear()%1e4,t,4)}function $V(e){var t=e.getTimezoneOffset();return(t>0?"-":(t*=-1,"+"))+he(t/60|0,"0",2)+he(t%60,"0",2)}function I3(e,t){return he(e.getUTCDate(),t,2)}function MV(e,t){return he(e.getUTCHours(),t,2)}function RV(e,t){return he(e.getUTCHours()%12||12,t,2)}function IV(e,t){return he(1+cd.count(zr(e),e),t,3)}function R7(e,t){return he(e.getUTCMilliseconds(),t,3)}function DV(e,t){return R7(e,t)+"000"}function LV(e,t){return he(e.getUTCMonth()+1,t,2)}function FV(e,t){return he(e.getUTCMinutes(),t,2)}function BV(e,t){return he(e.getUTCSeconds(),t,2)}function zV(e){var t=e.getUTCDay();return t===0?7:t}function HV(e,t){return he(fd.count(zr(e)-1,e),t,2)}function I7(e){var t=e.getUTCDay();return t>=4||t===0?Li(e):Li.ceil(e)}function GV(e,t){return e=I7(e),he(Li.count(zr(e),e)+(zr(e).getUTCDay()===4),t,2)}function UV(e){return e.getUTCDay()}function WV(e,t){return he(Bp.count(zr(e)-1,e),t,2)}function qV(e,t){return he(e.getUTCFullYear()%100,t,2)}function VV(e,t){return e=I7(e),he(e.getUTCFullYear()%100,t,2)}function KV(e,t){return he(e.getUTCFullYear()%1e4,t,4)}function XV(e,t){var n=e.getUTCDay();return e=n>=4||n===0?Li(e):Li.ceil(e),he(e.getUTCFullYear()%1e4,t,4)}function YV(){return"+0000"}function D3(){return"%"}function L3(e){return+e}function F3(e){return Math.floor(+e/1e3)}var Io,D7,L7;QV({dateTime:"%x, %X",date:"%-m/%-d/%Y",time:"%-I:%M:%S %p",periods:["AM","PM"],days:["Sunday","Monday","Tuesday","Wednesday","Thursday","Friday","Saturday"],shortDays:["Sun","Mon","Tue","Wed","Thu","Fri","Sat"],months:["January","February","March","April","May","June","July","August","September","October","November","December"],shortMonths:["Jan","Feb","Mar","Apr","May","Jun","Jul","Aug","Sep","Oct","Nov","Dec"]});function QV(e){return Io=Qq(e),D7=Io.format,Io.parse,L7=Io.utcFormat,Io.utcParse,Io}function ZV(e){return new Date(e)}function JV(e){return e instanceof Date?+e:+new Date(+e)}function Pg(e,t,n,r,a,o,i,s,l,u){var p=lg(),c=p.invert,f=p.domain,m=u(".%L"),h=u(":%S"),d=u("%I:%M"),g=u("%I %p"),v=u("%a %d"),y=u("%b %d"),x=u("%B"),P=u("%Y");function w(S){return(l(S)<S?m:s(S)<S?h:i(S)<S?d:o(S)<S?g:r(S)<S?a(S)<S?v:y:n(S)<S?x:P)(S)}return p.invert=function(S){return new Date(c(S))},p.domain=function(S){return arguments.length?f(Array.from(S,JV)):f().map(ZV)},p.ticks=function(S){var O=f();return e(O[0],O[O.length-1],S??10)},p.tickFormat=function(S,O){return O==null?w:u(O)},p.nice=function(S){var O=f();return(!S||typeof S.range!="function")&&(S=t(O[0],O[O.length-1],S??10)),S?f(O7(O,S)):p},p.copy=function(){return Au(p,Pg(e,t,n,r,a,o,i,s,l,u))},p}function eK(){return Pn.apply(Pg(Xq,Yq,Br,wg,pd,Eu,gg,vg,Qa,D7).domain([new Date(2e3,0,1),new Date(2e3,0,2)]),arguments)}function tK(){return Pn.apply(Pg(Vq,Kq,zr,bg,fd,cd,xg,yg,Qa,L7).domain([Date.UTC(2e3,0,1),Date.UTC(2e3,0,2)]),arguments)}function dd(){var e=0,t=1,n,r,a,o,i=$t,s=!1,l;function u(c){return c==null||isNaN(c=+c)?l:i(a===0?.5:(c=(o(c)-n)*a,s?Math.max(0,Math.min(1,c)):c))}u.domain=function(c){return arguments.length?([e,t]=c,n=o(e=+e),r=o(t=+t),a=n===r?0:1/(r-n),u):[e,t]},u.clamp=function(c){return arguments.length?(s=!!c,u):s},u.interpolator=function(c){return arguments.length?(i=c,u):i};function p(c){return function(f){var m,h;return arguments.length?([m,h]=f,i=c(m,h),u):[i(0),i(1)]}}return u.range=p(bs),u.rangeRound=p(sg),u.unknown=function(c){return arguments.length?(l=c,u):l},function(c){return o=c,n=c(e),r=c(t),a=n===r?0:1/(r-n),u}}function Ra(e,t){return t.domain(e.domain()).interpolator(e.interpolator()).clamp(e.clamp()).unknown(e.unknown())}function F7(){var e=Ma(dd()($t));return e.copy=function(){return Ra(e,F7())},Wr.apply(e,arguments)}function B7(){var e=pg(dd()).domain([1,10]);return e.copy=function(){return Ra(e,B7()).base(e.base())},Wr.apply(e,arguments)}function z7(){var e=fg(dd());return e.copy=function(){return Ra(e,z7()).constant(e.constant())},Wr.apply(e,arguments)}function Sg(){var e=dg(dd());return e.copy=function(){return Ra(e,Sg()).exponent(e.exponent())},Wr.apply(e,arguments)}function nK(){return Sg.apply(null,arguments).exponent(.5)}function H7(){var e=[],t=$t;function n(r){if(r!=null&&!isNaN(r=+r))return t((Cu(e,r,1)-1)/(e.length-1))}return n.domain=function(r){if(!arguments.length)return e.slice();e=[];for(let a of r)a!=null&&!isNaN(a=+a)&&e.push(a);return e.sort(ba),n},n.interpolator=function(r){return arguments.length?(t=r,n):t},n.range=function(){return e.map((r,a)=>t(a/(e.length-1)))},n.quantiles=function(r){return Array.from({length:r+1},(a,o)=>HW(e,o/r))},n.copy=function(){return H7(t).domain(e)},Wr.apply(n,arguments)}function md(){var e=0,t=.5,n=1,r=1,a,o,i,s,l,u=$t,p,c=!1,f;function m(d){return isNaN(d=+d)?f:(d=.5+((d=+p(d))-o)*(r*d<r*o?s:l),u(c?Math.max(0,Math.min(1,d)):d))}m.domain=function(d){return arguments.length?([e,t,n]=d,a=p(e=+e),o=p(t=+t),i=p(n=+n),s=a===o?0:.5/(o-a),l=o===i?0:.5/(i-o),r=o<a?-1:1,m):[e,t,n]},m.clamp=function(d){return arguments.length?(c=!!d,m):c},m.interpolator=function(d){return arguments.length?(u=d,m):u};function h(d){return function(g){var v,y,x;return arguments.length?([v,y,x]=g,u=mq(d,[v,y,x]),m):[u(0),u(.5),u(1)]}}return m.range=h(bs),m.rangeRound=h(sg),m.unknown=function(d){return arguments.length?(f=d,m):f},function(d){return p=d,a=d(e),o=d(t),i=d(n),s=a===o?0:.5/(o-a),l=o===i?0:.5/(i-o),r=o<a?-1:1,m}}function G7(){var e=Ma(md()($t));return e.copy=function(){return Ra(e,G7())},Wr.apply(e,arguments)}function U7(){var e=pg(md()).domain([.1,1,10]);return e.copy=function(){return Ra(e,U7()).base(e.base())},Wr.apply(e,arguments)}function W7(){var e=fg(md());return e.copy=function(){return Ra(e,W7()).constant(e.constant())},Wr.apply(e,arguments)}function Og(){var e=dg(md());return e.copy=function(){return Ra(e,Og()).exponent(e.exponent())},Wr.apply(e,arguments)}function rK(){return Og.apply(null,arguments).exponent(.5)}const B3=Object.freeze(Object.defineProperty({__proto__:null,scaleBand:Ll,scaleDiverging:G7,scaleDivergingLog:U7,scaleDivergingPow:Og,scaleDivergingSqrt:rK,scaleDivergingSymlog:W7,scaleIdentity:S7,scaleImplicit:k0,scaleLinear:Dp,scaleLog:k7,scaleOrdinal:ag,scalePoint:sl,scalePow:mg,scaleQuantile:A7,scaleQuantize:E7,scaleRadial:_7,scaleSequential:F7,scaleSequentialLog:B7,scaleSequentialPow:Sg,scaleSequentialQuantile:H7,scaleSequentialSqrt:nK,scaleSequentialSymlog:z7,scaleSqrt:Dq,scaleSymlog:C7,scaleThreshold:T7,scaleTime:eK,scaleUtc:tK,tickFormat:P7},Symbol.toStringTag,{value:"Module"}));var aK=ps;function oK(e,t,n){for(var r=-1,a=e.length;++r<a;){var o=e[r],i=t(o);if(i!=null&&(s===void 0?i===i&&!aK(i):n(i,s)))var s=i,l=o}return l}var q7=oK;function iK(e,t){return e>t}var sK=iK,lK=q7,uK=sK,cK=ws;function pK(e){return e&&e.length?lK(e,cK,uK):void 0}var fK=pK;const pa=_e(fK);function dK(e,t){return e<t}var mK=dK,hK=q7,vK=mK,yK=ws;function gK(e){return e&&e.length?hK(e,yK,vK):void 0}var xK=gK;const hd=_e(xK);var wK=Ly,bK=$a,PK=e7,SK=qt;function OK(e,t){var n=SK(e)?wK:PK;return n(e,bK(t))}var kK=OK,CK=Z8,_K=kK;function AK(e,t){return CK(_K(e,t),1)}var EK=AK;const TK=_e(EK);var jK=Jy;function NK(e,t){return jK(e,t)}var $K=NK;const Fi=_e($K);var Ps=1e9,MK={precision:20,rounding:4,toExpNeg:-7,toExpPos:21,LN10:"2.302585092994045684017991454684364207601101488628772976033327900967572609677352480235997205089598298341967784042286"},Cg,ze=!0,wn="[DecimalError] ",uo=wn+"Invalid argument: ",kg=wn+"Exponent out of range: ",Ss=Math.floor,Wa=Math.pow,RK=/^(\d+(\.\d*)?|\.\d+)(e[+-]?\d+)?$/i,Zt,st=1e7,Le=7,V7=9007199254740991,zp=Ss(V7/Le),Q={};Q.absoluteValue=Q.abs=function(){var e=new this.constructor(this);return e.s&&(e.s=1),e};Q.comparedTo=Q.cmp=function(e){var t,n,r,a,o=this;if(e=new o.constructor(e),o.s!==e.s)return o.s||-e.s;if(o.e!==e.e)return o.e>e.e^o.s<0?1:-1;for(r=o.d.length,a=e.d.length,t=0,n=r<a?r:a;t<n;++t)if(o.d[t]!==e.d[t])return o.d[t]>e.d[t]^o.s<0?1:-1;return r===a?0:r>a^o.s<0?1:-1};Q.decimalPlaces=Q.dp=function(){var e=this,t=e.d.length-1,n=(t-e.e)*Le;if(t=e.d[t],t)for(;t%10==0;t/=10)n--;return n<0?0:n};Q.dividedBy=Q.div=function(e){return jr(this,new this.constructor(e))};Q.dividedToIntegerBy=Q.idiv=function(e){var t=this,n=t.constructor;return Ee(jr(t,new n(e),0,1),n.precision)};Q.equals=Q.eq=function(e){return!this.cmp(e)};Q.exponent=function(){return et(this)};Q.greaterThan=Q.gt=function(e){return this.cmp(e)>0};Q.greaterThanOrEqualTo=Q.gte=function(e){return this.cmp(e)>=0};Q.isInteger=Q.isint=function(){return this.e>this.d.length-2};Q.isNegative=Q.isneg=function(){return this.s<0};Q.isPositive=Q.ispos=function(){return this.s>0};Q.isZero=function(){return this.s===0};Q.lessThan=Q.lt=function(e){return this.cmp(e)<0};Q.lessThanOrEqualTo=Q.lte=function(e){return this.cmp(e)<1};Q.logarithm=Q.log=function(e){var t,n=this,r=n.constructor,a=r.precision,o=a+5;if(e===void 0)e=new r(10);else if(e=new r(e),e.s<1||e.eq(Zt))throw Error(wn+"NaN");if(n.s<1)throw Error(wn+(n.s?"NaN":"-Infinity"));return n.eq(Zt)?new r(0):(ze=!1,t=jr(Gl(n,o),Gl(e,o),o),ze=!0,Ee(t,a))};Q.minus=Q.sub=function(e){var t=this;return e=new t.constructor(e),t.s==e.s?Y7(t,e):K7(t,(e.s=-e.s,e))};Q.modulo=Q.mod=function(e){var t,n=this,r=n.constructor,a=r.precision;if(e=new r(e),!e.s)throw Error(wn+"NaN");return n.s?(ze=!1,t=jr(n,e,0,1).times(e),ze=!0,n.minus(t)):Ee(new r(n),a)};Q.naturalExponential=Q.exp=function(){return X7(this)};Q.naturalLogarithm=Q.ln=function(){return Gl(this)};Q.negated=Q.neg=function(){var e=new this.constructor(this);return e.s=-e.s||0,e};Q.plus=Q.add=function(e){var t=this;return e=new t.constructor(e),t.s==e.s?K7(t,e):Y7(t,(e.s=-e.s,e))};Q.precision=Q.sd=function(e){var t,n,r,a=this;if(e!==void 0&&e!==!!e&&e!==1&&e!==0)throw Error(uo+e);if(t=et(a)+1,r=a.d.length-1,n=r*Le+1,r=a.d[r],r){for(;r%10==0;r/=10)n--;for(r=a.d[0];r>=10;r/=10)n++}return e&&t>n?t:n};Q.squareRoot=Q.sqrt=function(){var e,t,n,r,a,o,i,s=this,l=s.constructor;if(s.s<1){if(!s.s)return new l(0);throw Error(wn+"NaN")}for(e=et(s),ze=!1,a=Math.sqrt(+s),a==0||a==1/0?(t=rr(s.d),(t.length+e)%2==0&&(t+="0"),a=Math.sqrt(t),e=Ss((e+1)/2)-(e<0||e%2),a==1/0?t="5e"+e:(t=a.toExponential(),t=t.slice(0,t.indexOf("e")+1)+e),r=new l(t)):r=new l(a.toString()),n=l.precision,a=i=n+3;;)if(o=r,r=o.plus(jr(s,o,i+2)).times(.5),rr(o.d).slice(0,i)===(t=rr(r.d)).slice(0,i)){if(t=t.slice(i-3,i+1),a==i&&t=="4999"){if(Ee(o,n+1,0),o.times(o).eq(s)){r=o;break}}else if(t!="9999")break;i+=4}return ze=!0,Ee(r,n)};Q.times=Q.mul=function(e){var t,n,r,a,o,i,s,l,u,p=this,c=p.constructor,f=p.d,m=(e=new c(e)).d;if(!p.s||!e.s)return new c(0);for(e.s*=p.s,n=p.e+e.e,l=f.length,u=m.length,l<u&&(o=f,f=m,m=o,i=l,l=u,u=i),o=[],i=l+u,r=i;r--;)o.push(0);for(r=u;--r>=0;){for(t=0,a=l+r;a>r;)s=o[a]+m[r]*f[a-r-1]+t,o[a--]=s%st|0,t=s/st|0;o[a]=(o[a]+t)%st|0}for(;!o[--i];)o.pop();return t?++n:o.shift(),e.d=o,e.e=n,ze?Ee(e,c.precision):e};Q.toDecimalPlaces=Q.todp=function(e,t){var n=this,r=n.constructor;return n=new r(n),e===void 0?n:(dr(e,0,Ps),t===void 0?t=r.rounding:dr(t,0,8),Ee(n,e+et(n)+1,t))};Q.toExponential=function(e,t){var n,r=this,a=r.constructor;return e===void 0?n=bo(r,!0):(dr(e,0,Ps),t===void 0?t=a.rounding:dr(t,0,8),r=Ee(new a(r),e+1,t),n=bo(r,!0,e+1)),n};Q.toFixed=function(e,t){var n,r,a=this,o=a.constructor;return e===void 0?bo(a):(dr(e,0,Ps),t===void 0?t=o.rounding:dr(t,0,8),r=Ee(new o(a),e+et(a)+1,t),n=bo(r.abs(),!1,e+et(r)+1),a.isneg()&&!a.isZero()?"-"+n:n)};Q.toInteger=Q.toint=function(){var e=this,t=e.constructor;return Ee(new t(e),et(e)+1,t.rounding)};Q.toNumber=function(){return+this};Q.toPower=Q.pow=function(e){var t,n,r,a,o,i,s=this,l=s.constructor,u=12,p=+(e=new l(e));if(!e.s)return new l(Zt);if(s=new l(s),!s.s){if(e.s<1)throw Error(wn+"Infinity");return s}if(s.eq(Zt))return s;if(r=l.precision,e.eq(Zt))return Ee(s,r);if(t=e.e,n=e.d.length-1,i=t>=n,o=s.s,i){if((n=p<0?-p:p)<=V7){for(a=new l(Zt),t=Math.ceil(r/Le+4),ze=!1;n%2&&(a=a.times(s),H3(a.d,t)),n=Ss(n/2),n!==0;)s=s.times(s),H3(s.d,t);return ze=!0,e.s<0?new l(Zt).div(a):Ee(a,r)}}else if(o<0)throw Error(wn+"NaN");return o=o<0&&e.d[Math.max(t,n)]&1?-1:1,s.s=1,ze=!1,a=e.times(Gl(s,r+u)),ze=!0,a=X7(a),a.s=o,a};Q.toPrecision=function(e,t){var n,r,a=this,o=a.constructor;return e===void 0?(n=et(a),r=bo(a,n<=o.toExpNeg||n>=o.toExpPos)):(dr(e,1,Ps),t===void 0?t=o.rounding:dr(t,0,8),a=Ee(new o(a),e,t),n=et(a),r=bo(a,e<=n||n<=o.toExpNeg,e)),r};Q.toSignificantDigits=Q.tosd=function(e,t){var n=this,r=n.constructor;return e===void 0?(e=r.precision,t=r.rounding):(dr(e,1,Ps),t===void 0?t=r.rounding:dr(t,0,8)),Ee(new r(n),e,t)};Q.toString=Q.valueOf=Q.val=Q.toJSON=Q[Symbol.for("nodejs.util.inspect.custom")]=function(){var e=this,t=et(e),n=e.constructor;return bo(e,t<=n.toExpNeg||t>=n.toExpPos)};function K7(e,t){var n,r,a,o,i,s,l,u,p=e.constructor,c=p.precision;if(!e.s||!t.s)return t.s||(t=new p(e)),ze?Ee(t,c):t;if(l=e.d,u=t.d,i=e.e,a=t.e,l=l.slice(),o=i-a,o){for(o<0?(r=l,o=-o,s=u.length):(r=u,a=i,s=l.length),i=Math.ceil(c/Le),s=i>s?i+1:s+1,o>s&&(o=s,r.length=1),r.reverse();o--;)r.push(0);r.reverse()}for(s=l.length,o=u.length,s-o<0&&(o=s,r=u,u=l,l=r),n=0;o;)n=(l[--o]=l[o]+u[o]+n)/st|0,l[o]%=st;for(n&&(l.unshift(n),++a),s=l.length;l[--s]==0;)l.pop();return t.d=l,t.e=a,ze?Ee(t,c):t}function dr(e,t,n){if(e!==~~e||e<t||e>n)throw Error(uo+e)}function rr(e){var t,n,r,a=e.length-1,o="",i=e[0];if(a>0){for(o+=i,t=1;t<a;t++)r=e[t]+"",n=Le-r.length,n&&(o+=ea(n)),o+=r;i=e[t],r=i+"",n=Le-r.length,n&&(o+=ea(n))}else if(i===0)return"0";for(;i%10===0;)i/=10;return o+i}var jr=function(){function e(r,a){var o,i=0,s=r.length;for(r=r.slice();s--;)o=r[s]*a+i,r[s]=o%st|0,i=o/st|0;return i&&r.unshift(i),r}function t(r,a,o,i){var s,l;if(o!=i)l=o>i?1:-1;else for(s=l=0;s<o;s++)if(r[s]!=a[s]){l=r[s]>a[s]?1:-1;break}return l}function n(r,a,o){for(var i=0;o--;)r[o]-=i,i=r[o]<a[o]?1:0,r[o]=i*st+r[o]-a[o];for(;!r[0]&&r.length>1;)r.shift()}return function(r,a,o,i){var s,l,u,p,c,f,m,h,d,g,v,y,x,P,w,S,O,C,_=r.constructor,T=r.s==a.s?1:-1,A=r.d,j=a.d;if(!r.s)return new _(r);if(!a.s)throw Error(wn+"Division by zero");for(l=r.e-a.e,O=j.length,w=A.length,m=new _(T),h=m.d=[],u=0;j[u]==(A[u]||0);)++u;if(j[u]>(A[u]||0)&&--l,o==null?y=o=_.precision:i?y=o+(et(r)-et(a))+1:y=o,y<0)return new _(0);if(y=y/Le+2|0,u=0,O==1)for(p=0,j=j[0],y++;(u<w||p)&&y--;u++)x=p*st+(A[u]||0),h[u]=x/j|0,p=x%j|0;else{for(p=st/(j[0]+1)|0,p>1&&(j=e(j,p),A=e(A,p),O=j.length,w=A.length),P=O,d=A.slice(0,O),g=d.length;g<O;)d[g++]=0;C=j.slice(),C.unshift(0),S=j[0],j[1]>=st/2&&++S;do p=0,s=t(j,d,O,g),s<0?(v=d[0],O!=g&&(v=v*st+(d[1]||0)),p=v/S|0,p>1?(p>=st&&(p=st-1),c=e(j,p),f=c.length,g=d.length,s=t(c,d,f,g),s==1&&(p--,n(c,O<f?C:j,f))):(p==0&&(s=p=1),c=j.slice()),f=c.length,f<g&&c.unshift(0),n(d,c,g),s==-1&&(g=d.length,s=t(j,d,O,g),s<1&&(p++,n(d,O<g?C:j,g))),g=d.length):s===0&&(p++,d=[0]),h[u++]=p,s&&d[0]?d[g++]=A[P]||0:(d=[A[P]],g=1);while((P++<w||d[0]!==void 0)&&y--)}return h[0]||h.shift(),m.e=l,Ee(m,i?o+et(m)+1:o)}}();function X7(e,t){var n,r,a,o,i,s,l=0,u=0,p=e.constructor,c=p.precision;if(et(e)>16)throw Error(kg+et(e));if(!e.s)return new p(Zt);for(ze=!1,s=c,i=new p(.03125);e.abs().gte(.1);)e=e.times(i),u+=5;for(r=Math.log(Wa(2,u))/Math.LN10*2+5|0,s+=r,n=a=o=new p(Zt),p.precision=s;;){if(a=Ee(a.times(e),s),n=n.times(++l),i=o.plus(jr(a,n,s)),rr(i.d).slice(0,s)===rr(o.d).slice(0,s)){for(;u--;)o=Ee(o.times(o),s);return p.precision=c,t==null?(ze=!0,Ee(o,c)):o}o=i}}function et(e){for(var t=e.e*Le,n=e.d[0];n>=10;n/=10)t++;return t}function Rm(e,t,n){if(t>e.LN10.sd())throw ze=!0,n&&(e.precision=n),Error(wn+"LN10 precision limit exceeded");return Ee(new e(e.LN10),t)}function ea(e){for(var t="";e--;)t+="0";return t}function Gl(e,t){var n,r,a,o,i,s,l,u,p,c=1,f=10,m=e,h=m.d,d=m.constructor,g=d.precision;if(m.s<1)throw Error(wn+(m.s?"NaN":"-Infinity"));if(m.eq(Zt))return new d(0);if(t==null?(ze=!1,u=g):u=t,m.eq(10))return t==null&&(ze=!0),Rm(d,u);if(u+=f,d.precision=u,n=rr(h),r=n.charAt(0),o=et(m),Math.abs(o)<15e14){for(;r<7&&r!=1||r==1&&n.charAt(1)>3;)m=m.times(e),n=rr(m.d),r=n.charAt(0),c++;o=et(m),r>1?(m=new d("0."+n),o++):m=new d(r+"."+n.slice(1))}else return l=Rm(d,u+2,g).times(o+""),m=Gl(new d(r+"."+n.slice(1)),u-f).plus(l),d.precision=g,t==null?(ze=!0,Ee(m,g)):m;for(s=i=m=jr(m.minus(Zt),m.plus(Zt),u),p=Ee(m.times(m),u),a=3;;){if(i=Ee(i.times(p),u),l=s.plus(jr(i,new d(a),u)),rr(l.d).slice(0,u)===rr(s.d).slice(0,u))return s=s.times(2),o!==0&&(s=s.plus(Rm(d,u+2,g).times(o+""))),s=jr(s,new d(c),u),d.precision=g,t==null?(ze=!0,Ee(s,g)):s;s=l,a+=2}}function z3(e,t){var n,r,a;for((n=t.indexOf("."))>-1&&(t=t.replace(".","")),(r=t.search(/e/i))>0?(n<0&&(n=r),n+=+t.slice(r+1),t=t.substring(0,r)):n<0&&(n=t.length),r=0;t.charCodeAt(r)===48;)++r;for(a=t.length;t.charCodeAt(a-1)===48;)--a;if(t=t.slice(r,a),t){if(a-=r,n=n-r-1,e.e=Ss(n/Le),e.d=[],r=(n+1)%Le,n<0&&(r+=Le),r<a){for(r&&e.d.push(+t.slice(0,r)),a-=Le;r<a;)e.d.push(+t.slice(r,r+=Le));t=t.slice(r),r=Le-t.length}else r-=a;for(;r--;)t+="0";if(e.d.push(+t),ze&&(e.e>zp||e.e<-zp))throw Error(kg+n)}else e.s=0,e.e=0,e.d=[0];return e}function Ee(e,t,n){var r,a,o,i,s,l,u,p,c=e.d;for(i=1,o=c[0];o>=10;o/=10)i++;if(r=t-i,r<0)r+=Le,a=t,u=c[p=0];else{if(p=Math.ceil((r+1)/Le),o=c.length,p>=o)return e;for(u=o=c[p],i=1;o>=10;o/=10)i++;r%=Le,a=r-Le+i}if(n!==void 0&&(o=Wa(10,i-a-1),s=u/o%10|0,l=t<0||c[p+1]!==void 0||u%o,l=n<4?(s||l)&&(n==0||n==(e.s<0?3:2)):s>5||s==5&&(n==4||l||n==6&&(r>0?a>0?u/Wa(10,i-a):0:c[p-1])%10&1||n==(e.s<0?8:7))),t<1||!c[0])return l?(o=et(e),c.length=1,t=t-o-1,c[0]=Wa(10,(Le-t%Le)%Le),e.e=Ss(-t/Le)||0):(c.length=1,c[0]=e.e=e.s=0),e;if(r==0?(c.length=p,o=1,p--):(c.length=p+1,o=Wa(10,Le-r),c[p]=a>0?(u/Wa(10,i-a)%Wa(10,a)|0)*o:0),l)for(;;)if(p==0){(c[0]+=o)==st&&(c[0]=1,++e.e);break}else{if(c[p]+=o,c[p]!=st)break;c[p--]=0,o=1}for(r=c.length;c[--r]===0;)c.pop();if(ze&&(e.e>zp||e.e<-zp))throw Error(kg+et(e));return e}function Y7(e,t){var n,r,a,o,i,s,l,u,p,c,f=e.constructor,m=f.precision;if(!e.s||!t.s)return t.s?t.s=-t.s:t=new f(e),ze?Ee(t,m):t;if(l=e.d,c=t.d,r=t.e,u=e.e,l=l.slice(),i=u-r,i){for(p=i<0,p?(n=l,i=-i,s=c.length):(n=c,r=u,s=l.length),a=Math.max(Math.ceil(m/Le),s)+2,i>a&&(i=a,n.length=1),n.reverse(),a=i;a--;)n.push(0);n.reverse()}else{for(a=l.length,s=c.length,p=a<s,p&&(s=a),a=0;a<s;a++)if(l[a]!=c[a]){p=l[a]<c[a];break}i=0}for(p&&(n=l,l=c,c=n,t.s=-t.s),s=l.length,a=c.length-s;a>0;--a)l[s++]=0;for(a=c.length;a>i;){if(l[--a]<c[a]){for(o=a;o&&l[--o]===0;)l[o]=st-1;--l[o],l[a]+=st}l[a]-=c[a]}for(;l[--s]===0;)l.pop();for(;l[0]===0;l.shift())--r;return l[0]?(t.d=l,t.e=r,ze?Ee(t,m):t):new f(0)}function bo(e,t,n){var r,a=et(e),o=rr(e.d),i=o.length;return t?(n&&(r=n-i)>0?o=o.charAt(0)+"."+o.slice(1)+ea(r):i>1&&(o=o.charAt(0)+"."+o.slice(1)),o=o+(a<0?"e":"e+")+a):a<0?(o="0."+ea(-a-1)+o,n&&(r=n-i)>0&&(o+=ea(r))):a>=i?(o+=ea(a+1-i),n&&(r=n-a-1)>0&&(o=o+"."+ea(r))):((r=a+1)<i&&(o=o.slice(0,r)+"."+o.slice(r)),n&&(r=n-i)>0&&(a+1===i&&(o+="."),o+=ea(r))),e.s<0?"-"+o:o}function H3(e,t){if(e.length>t)return e.length=t,!0}function Q7(e){var t,n,r;function a(o){var i=this;if(!(i instanceof a))return new a(o);if(i.constructor=a,o instanceof a){i.s=o.s,i.e=o.e,i.d=(o=o.d)?o.slice():o;return}if(typeof o=="number"){if(o*0!==0)throw Error(uo+o);if(o>0)i.s=1;else if(o<0)o=-o,i.s=-1;else{i.s=0,i.e=0,i.d=[0];return}if(o===~~o&&o<1e7){i.e=0,i.d=[o];return}return z3(i,o.toString())}else if(typeof o!="string")throw Error(uo+o);if(o.charCodeAt(0)===45?(o=o.slice(1),i.s=-1):i.s=1,RK.test(o))z3(i,o);else throw Error(uo+o)}if(a.prototype=Q,a.ROUND_UP=0,a.ROUND_DOWN=1,a.ROUND_CEIL=2,a.ROUND_FLOOR=3,a.ROUND_HALF_UP=4,a.ROUND_HALF_DOWN=5,a.ROUND_HALF_EVEN=6,a.ROUND_HALF_CEIL=7,a.ROUND_HALF_FLOOR=8,a.clone=Q7,a.config=a.set=IK,e===void 0&&(e={}),e)for(r=["precision","rounding","toExpNeg","toExpPos","LN10"],t=0;t<r.length;)e.hasOwnProperty(n=r[t++])||(e[n]=this[n]);return a.config(e),a}function IK(e){if(!e||typeof e!="object")throw Error(wn+"Object expected");var t,n,r,a=["precision",1,Ps,"rounding",0,8,"toExpNeg",-1/0,0,"toExpPos",0,1/0];for(t=0;t<a.length;t+=3)if((r=e[n=a[t]])!==void 0)if(Ss(r)===r&&r>=a[t+1]&&r<=a[t+2])this[n]=r;else throw Error(uo+n+": "+r);if((r=e[n="LN10"])!==void 0)if(r==Math.LN10)this[n]=new this(r);else throw Error(uo+n+": "+r);return this}var Cg=Q7(MK);Zt=new Cg(1);const Ce=Cg;function DK(e){return zK(e)||BK(e)||FK(e)||LK()}function LK(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function FK(e,t){if(e){if(typeof e=="string")return E0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return E0(e,t)}}function BK(e){if(typeof Symbol<"u"&&Symbol.iterator in Object(e))return Array.from(e)}function zK(e){if(Array.isArray(e))return E0(e)}function E0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}var HK=function(t){return t},Z7={},J7=function(t){return t===Z7},G3=function(t){return function n(){return arguments.length===0||arguments.length===1&&J7(arguments.length<=0?void 0:arguments[0])?n:t.apply(void 0,arguments)}},GK=function e(t,n){return t===1?n:G3(function(){for(var r=arguments.length,a=new Array(r),o=0;o<r;o++)a[o]=arguments[o];var i=a.filter(function(s){return s!==Z7}).length;return i>=t?n.apply(void 0,a):e(t-i,G3(function(){for(var s=arguments.length,l=new Array(s),u=0;u<s;u++)l[u]=arguments[u];var p=a.map(function(c){return J7(c)?l.shift():c});return n.apply(void 0,DK(p).concat(l))}))})},vd=function(t){return GK(t.length,t)},T0=function(t,n){for(var r=[],a=t;a<n;++a)r[a-t]=a;return r},UK=vd(function(e,t){return Array.isArray(t)?t.map(e):Object.keys(t).map(function(n){return t[n]}).map(e)}),WK=function(){for(var t=arguments.length,n=new Array(t),r=0;r<t;r++)n[r]=arguments[r];if(!n.length)return HK;var a=n.reverse(),o=a[0],i=a.slice(1);return function(){return i.reduce(function(s,l){return l(s)},o.apply(void 0,arguments))}},j0=function(t){return Array.isArray(t)?t.reverse():t.split("").reverse.join("")},eP=function(t){var n=null,r=null;return function(){for(var a=arguments.length,o=new Array(a),i=0;i<a;i++)o[i]=arguments[i];return n&&o.every(function(s,l){return s===n[l]})||(n=o,r=t.apply(void 0,o)),r}};function qK(e){var t;return e===0?t=1:t=Math.floor(new Ce(e).abs().log(10).toNumber())+1,t}function VK(e,t,n){for(var r=new Ce(e),a=0,o=[];r.lt(t)&&a<1e5;)o.push(r.toNumber()),r=r.add(n),a++;return o}var KK=vd(function(e,t,n){var r=+e,a=+t;return r+n*(a-r)}),XK=vd(function(e,t,n){var r=t-+e;return r=r||1/0,(n-e)/r}),YK=vd(function(e,t,n){var r=t-+e;return r=r||1/0,Math.max(0,Math.min(1,(n-e)/r))});const yd={rangeStep:VK,getDigitCount:qK,interpolateNumber:KK,uninterpolateNumber:XK,uninterpolateTruncation:YK};function N0(e){return JK(e)||ZK(e)||tP(e)||QK()}function QK(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function ZK(e){if(typeof Symbol<"u"&&Symbol.iterator in Object(e))return Array.from(e)}function JK(e){if(Array.isArray(e))return $0(e)}function Ul(e,t){return nX(e)||tX(e,t)||tP(e,t)||eX()}function eX(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function tP(e,t){if(e){if(typeof e=="string")return $0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return $0(e,t)}}function $0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function tX(e,t){if(!(typeof Symbol>"u"||!(Symbol.iterator in Object(e)))){var n=[],r=!0,a=!1,o=void 0;try{for(var i=e[Symbol.iterator](),s;!(r=(s=i.next()).done)&&(n.push(s.value),!(t&&n.length===t));r=!0);}catch(l){a=!0,o=l}finally{try{!r&&i.return!=null&&i.return()}finally{if(a)throw o}}return n}}function nX(e){if(Array.isArray(e))return e}function nP(e){var t=Ul(e,2),n=t[0],r=t[1],a=n,o=r;return n>r&&(a=r,o=n),[a,o]}function rP(e,t,n){if(e.lte(0))return new Ce(0);var r=yd.getDigitCount(e.toNumber()),a=new Ce(10).pow(r),o=e.div(a),i=r!==1?.05:.1,s=new Ce(Math.ceil(o.div(i).toNumber())).add(n).mul(i),l=s.mul(a);return t?l:new Ce(Math.ceil(l))}function rX(e,t,n){var r=1,a=new Ce(e);if(!a.isint()&&n){var o=Math.abs(e);o<1?(r=new Ce(10).pow(yd.getDigitCount(e)-1),a=new Ce(Math.floor(a.div(r).toNumber())).mul(r)):o>1&&(a=new Ce(Math.floor(e)))}else e===0?a=new Ce(Math.floor((t-1)/2)):n||(a=new Ce(Math.floor(e)));var i=Math.floor((t-1)/2),s=WK(UK(function(l){return a.add(new Ce(l-i).mul(r)).toNumber()}),T0);return s(0,t)}function aP(e,t,n,r){var a=arguments.length>4&&arguments[4]!==void 0?arguments[4]:0;if(!Number.isFinite((t-e)/(n-1)))return{step:new Ce(0),tickMin:new Ce(0),tickMax:new Ce(0)};var o=rP(new Ce(t).sub(e).div(n-1),r,a),i;e<=0&&t>=0?i=new Ce(0):(i=new Ce(e).add(t).div(2),i=i.sub(new Ce(i).mod(o)));var s=Math.ceil(i.sub(e).div(o).toNumber()),l=Math.ceil(new Ce(t).sub(i).div(o).toNumber()),u=s+l+1;return u>n?aP(e,t,n,r,a+1):(u<n&&(l=t>0?l+(n-u):l,s=t>0?s:s+(n-u)),{step:o,tickMin:i.sub(new Ce(s).mul(o)),tickMax:i.add(new Ce(l).mul(o))})}function aX(e){var t=Ul(e,2),n=t[0],r=t[1],a=arguments.length>1&&arguments[1]!==void 0?arguments[1]:6,o=arguments.length>2&&arguments[2]!==void 0?arguments[2]:!0,i=Math.max(a,2),s=nP([n,r]),l=Ul(s,2),u=l[0],p=l[1];if(u===-1/0||p===1/0){var c=p===1/0?[u].concat(N0(T0(0,a-1).map(function(){return 1/0}))):[].concat(N0(T0(0,a-1).map(function(){return-1/0})),[p]);return n>r?j0(c):c}if(u===p)return rX(u,a,o);var f=aP(u,p,i,o),m=f.step,h=f.tickMin,d=f.tickMax,g=yd.rangeStep(h,d.add(new Ce(.1).mul(m)),m);return n>r?j0(g):g}function oX(e,t){var n=Ul(e,2),r=n[0],a=n[1],o=arguments.length>2&&arguments[2]!==void 0?arguments[2]:!0,i=nP([r,a]),s=Ul(i,2),l=s[0],u=s[1];if(l===-1/0||u===1/0)return[r,a];if(l===u)return[l];var p=Math.max(t,2),c=rP(new Ce(u).sub(l).div(p-1),o,0),f=[].concat(N0(yd.rangeStep(new Ce(l),new Ce(u).sub(new Ce(.99).mul(c)),c)),[u]);return r>a?j0(f):f}var iX=eP(aX),sX=eP(oX),lX="Invariant failed";function Po(e,t){throw new Error(lX)}var uX=["offset","layout","width","dataKey","data","dataPointFormatter","xAxis","yAxis"];function Bi(e){"@babel/helpers - typeof";return Bi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Bi(e)}function Hp(){return Hp=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Hp.apply(this,arguments)}function cX(e,t){return mX(e)||dX(e,t)||fX(e,t)||pX()}function pX(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function fX(e,t){if(e){if(typeof e=="string")return U3(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return U3(e,t)}}function U3(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function dX(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function mX(e){if(Array.isArray(e))return e}function hX(e,t){if(e==null)return{};var n=vX(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function vX(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function yX(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function gX(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,sP(r.key),r)}}function xX(e,t,n){return t&&gX(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function wX(e,t,n){return t=Gp(t),bX(e,oP()?Reflect.construct(t,n||[],Gp(e).constructor):t.apply(e,n))}function bX(e,t){if(t&&(Bi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return PX(e)}function PX(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function oP(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(oP=function(){return!!e})()}function Gp(e){return Gp=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},Gp(e)}function SX(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&M0(e,t)}function M0(e,t){return M0=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},M0(e,t)}function iP(e,t,n){return t=sP(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function sP(e){var t=OX(e,"string");return Bi(t)=="symbol"?t:t+""}function OX(e,t){if(Bi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Bi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var Tu=function(e){function t(){return yX(this,t),wX(this,t,arguments)}return SX(t,e),xX(t,[{key:"render",value:function(){var r=this.props,a=r.offset,o=r.layout,i=r.width,s=r.dataKey,l=r.data,u=r.dataPointFormatter,p=r.xAxis,c=r.yAxis,f=hX(r,uX),m=ie(f,!1);this.props.direction==="x"&&p.type!=="number"&&Po();var h=l.map(function(d){var g=u(d,s),v=g.x,y=g.y,x=g.value,P=g.errorVal;if(!P)return null;var w=[],S,O;if(Array.isArray(P)){var C=cX(P,2);S=C[0],O=C[1]}else S=O=P;if(o==="vertical"){var _=p.scale,T=y+a,A=T+i,j=T-i,$=_(x-S),M=_(x+O);w.push({x1:M,y1:A,x2:M,y2:j}),w.push({x1:$,y1:T,x2:M,y2:T}),w.push({x1:$,y1:A,x2:$,y2:j})}else if(o==="horizontal"){var I=c.scale,R=v+a,L=R-i,N=R+i,D=I(x-S),H=I(x+O);w.push({x1:L,y1:H,x2:N,y2:H}),w.push({x1:R,y1:D,x2:R,y2:H}),w.push({x1:L,y1:D,x2:N,y2:D})}return E.createElement(Ne,Hp({className:"recharts-errorBar",key:"bar-".concat(w.map(function(W){return"".concat(W.x1,"-").concat(W.x2,"-").concat(W.y1,"-").concat(W.y2)}))},m),w.map(function(W){return E.createElement("line",Hp({},W,{key:"line-".concat(W.x1,"-").concat(W.x2,"-").concat(W.y1,"-").concat(W.y2)}))}))});return E.createElement(Ne,{className:"recharts-errorBars"},h)}}])}(E.Component);iP(Tu,"defaultProps",{stroke:"black",strokeWidth:1.5,width:5,offset:0,layout:"horizontal"});iP(Tu,"displayName","ErrorBar");function Wl(e){"@babel/helpers - typeof";return Wl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Wl(e)}function W3(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Ba(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?W3(Object(n),!0).forEach(function(r){kX(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):W3(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function kX(e,t,n){return t=CX(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function CX(e){var t=_X(e,"string");return Wl(t)=="symbol"?t:t+""}function _X(e,t){if(Wl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Wl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var lP=function(t){var n=t.children,r=t.formattedGraphicalItems,a=t.legendWidth,o=t.legendContent,i=Yt(n,ci);if(!i)return null;var s=ci.defaultProps,l=s!==void 0?Ba(Ba({},s),i.props):{},u;return i.props&&i.props.payload?u=i.props&&i.props.payload:o==="children"?u=(r||[]).reduce(function(p,c){var f=c.item,m=c.props,h=m.sectors||m.data||[];return p.concat(h.map(function(d){return{type:i.props.iconType||f.props.legendType,value:d.name,color:d.fill,payload:d}}))},[]):u=(r||[]).map(function(p){var c=p.item,f=c.type.defaultProps,m=f!==void 0?Ba(Ba({},f),c.props):{},h=m.dataKey,d=m.name,g=m.legendType,v=m.hide;return{inactive:v,dataKey:h,type:l.iconType||g||"square",color:_g(c),value:d||h,payload:m}}),Ba(Ba(Ba({},l),ci.getWithHeight(i,a)),{},{payload:u,item:i})};function ql(e){"@babel/helpers - typeof";return ql=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},ql(e)}function q3(e){return jX(e)||TX(e)||EX(e)||AX()}function AX(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function EX(e,t){if(e){if(typeof e=="string")return R0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return R0(e,t)}}function TX(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function jX(e){if(Array.isArray(e))return R0(e)}function R0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function V3(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Ve(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?V3(Object(n),!0).forEach(function(r){fi(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):V3(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function fi(e,t,n){return t=NX(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function NX(e){var t=$X(e,"string");return ql(t)=="symbol"?t:t+""}function $X(e,t){if(ql(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(ql(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function It(e,t,n){return le(e)||le(t)?n:ot(t)?vn(e,t,n):ae(t)?t(e):n}function ll(e,t,n,r){var a=TK(e,function(s){return It(s,t)});if(n==="number"){var o=a.filter(function(s){return V(s)||parseFloat(s)});return o.length?[hd(o),pa(o)]:[1/0,-1/0]}var i=r?a.filter(function(s){return!le(s)}):a;return i.map(function(s){return ot(s)||s instanceof Date?s:""})}var MX=function(t){var n,r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:[],a=arguments.length>2?arguments[2]:void 0,o=arguments.length>3?arguments[3]:void 0,i=-1,s=(n=r==null?void 0:r.length)!==null&&n!==void 0?n:0;if(s<=1)return 0;if(o&&o.axisType==="angleAxis"&&Math.abs(Math.abs(o.range[1]-o.range[0])-360)<=1e-6)for(var l=o.range,u=0;u<s;u++){var p=u>0?a[u-1].coordinate:a[s-1].coordinate,c=a[u].coordinate,f=u>=s-1?a[0].coordinate:a[u+1].coordinate,m=void 0;if(In(c-p)!==In(f-c)){var h=[];if(In(f-c)===In(l[1]-l[0])){m=f;var d=c+l[1]-l[0];h[0]=Math.min(d,(d+p)/2),h[1]=Math.max(d,(d+p)/2)}else{m=p;var g=f+l[1]-l[0];h[0]=Math.min(c,(g+c)/2),h[1]=Math.max(c,(g+c)/2)}var v=[Math.min(c,(m+c)/2),Math.max(c,(m+c)/2)];if(t>v[0]&&t<=v[1]||t>=h[0]&&t<=h[1]){i=a[u].index;break}}else{var y=Math.min(p,f),x=Math.max(p,f);if(t>(y+c)/2&&t<=(x+c)/2){i=a[u].index;break}}}else for(var P=0;P<s;P++)if(P===0&&t<=(r[P].coordinate+r[P+1].coordinate)/2||P>0&&P<s-1&&t>(r[P].coordinate+r[P-1].coordinate)/2&&t<=(r[P].coordinate+r[P+1].coordinate)/2||P===s-1&&t>(r[P].coordinate+r[P-1].coordinate)/2){i=r[P].index;break}return i},_g=function(t){var n,r=t,a=r.type.displayName,o=(n=t.type)!==null&&n!==void 0&&n.defaultProps?Ve(Ve({},t.type.defaultProps),t.props):t.props,i=o.stroke,s=o.fill,l;switch(a){case"Line":l=i;break;case"Area":case"Radar":l=i&&i!=="none"?i:s;break;default:l=s;break}return l},RX=function(t){var n=t.barSize,r=t.totalSize,a=t.stackGroups,o=a===void 0?{}:a;if(!o)return{};for(var i={},s=Object.keys(o),l=0,u=s.length;l<u;l++)for(var p=o[s[l]].stackGroups,c=Object.keys(p),f=0,m=c.length;f<m;f++){var h=p[c[f]],d=h.items,g=h.cateAxisId,v=d.filter(function(O){return Er(O.type).indexOf("Bar")>=0});if(v&&v.length){var y=v[0].type.defaultProps,x=y!==void 0?Ve(Ve({},y),v[0].props):v[0].props,P=x.barSize,w=x[g];i[w]||(i[w]=[]);var S=le(P)?n:P;i[w].push({item:v[0],stackList:v.slice(1),barSize:le(S)?void 0:wo(S,r,0)})}}return i},IX=function(t){var n=t.barGap,r=t.barCategoryGap,a=t.bandSize,o=t.sizeList,i=o===void 0?[]:o,s=t.maxBarSize,l=i.length;if(l<1)return null;var u=wo(n,a,0,!0),p,c=[];if(i[0].barSize===+i[0].barSize){var f=!1,m=a/l,h=i.reduce(function(P,w){return P+w.barSize||0},0);h+=(l-1)*u,h>=a&&(h-=(l-1)*u,u=0),h>=a&&m>0&&(f=!0,m*=.9,h=l*m);var d=(a-h)/2>>0,g={offset:d-u,size:0};p=i.reduce(function(P,w){var S={item:w.item,position:{offset:g.offset+g.size+u,size:f?m:w.barSize}},O=[].concat(q3(P),[S]);return g=O[O.length-1].position,w.stackList&&w.stackList.length&&w.stackList.forEach(function(C){O.push({item:C,position:g})}),O},c)}else{var v=wo(r,a,0,!0);a-2*v-(l-1)*u<=0&&(u=0);var y=(a-2*v-(l-1)*u)/l;y>1&&(y>>=0);var x=s===+s?Math.min(y,s):y;p=i.reduce(function(P,w,S){var O=[].concat(q3(P),[{item:w.item,position:{offset:v+(y+u)*S+(y-x)/2,size:x}}]);return w.stackList&&w.stackList.length&&w.stackList.forEach(function(C){O.push({item:C,position:O[O.length-1].position})}),O},c)}return p},DX=function(t,n,r,a){var o=r.children,i=r.width,s=r.margin,l=i-(s.left||0)-(s.right||0),u=lP({children:o,legendWidth:l});if(u){var p=a||{},c=p.width,f=p.height,m=u.align,h=u.verticalAlign,d=u.layout;if((d==="vertical"||d==="horizontal"&&h==="middle")&&m!=="center"&&V(t[m]))return Ve(Ve({},t),{},fi({},m,t[m]+(c||0)));if((d==="horizontal"||d==="vertical"&&m==="center")&&h!=="middle"&&V(t[h]))return Ve(Ve({},t),{},fi({},h,t[h]+(f||0)))}return t},LX=function(t,n,r){return le(n)?!0:t==="horizontal"?n==="yAxis":t==="vertical"||r==="x"?n==="xAxis":r==="y"?n==="yAxis":!0},uP=function(t,n,r,a,o){var i=n.props.children,s=yn(i,Tu).filter(function(u){return LX(a,o,u.props.direction)});if(s&&s.length){var l=s.map(function(u){return u.props.dataKey});return t.reduce(function(u,p){var c=It(p,r);if(le(c))return u;var f=Array.isArray(c)?[hd(c),pa(c)]:[c,c],m=l.reduce(function(h,d){var g=It(p,d,0),v=f[0]-Math.abs(Array.isArray(g)?g[0]:g),y=f[1]+Math.abs(Array.isArray(g)?g[1]:g);return[Math.min(v,h[0]),Math.max(y,h[1])]},[1/0,-1/0]);return[Math.min(m[0],u[0]),Math.max(m[1],u[1])]},[1/0,-1/0])}return null},FX=function(t,n,r,a,o){var i=n.map(function(s){return uP(t,s,r,o,a)}).filter(function(s){return!le(s)});return i&&i.length?i.reduce(function(s,l){return[Math.min(s[0],l[0]),Math.max(s[1],l[1])]},[1/0,-1/0]):null},cP=function(t,n,r,a,o){var i=n.map(function(l){var u=l.props.dataKey;return r==="number"&&u&&uP(t,l,u,a)||ll(t,u,r,o)});if(r==="number")return i.reduce(function(l,u){return[Math.min(l[0],u[0]),Math.max(l[1],u[1])]},[1/0,-1/0]);var s={};return i.reduce(function(l,u){for(var p=0,c=u.length;p<c;p++)s[u[p]]||(s[u[p]]=!0,l.push(u[p]));return l},[])},pP=function(t,n){return t==="horizontal"&&n==="xAxis"||t==="vertical"&&n==="yAxis"||t==="centric"&&n==="angleAxis"||t==="radial"&&n==="radiusAxis"},fP=function(t,n,r,a){if(a)return t.map(function(l){return l.coordinate});var o,i,s=t.map(function(l){return l.coordinate===n&&(o=!0),l.coordinate===r&&(i=!0),l.coordinate});return o||s.push(n),i||s.push(r),s},_r=function(t,n,r){if(!t)return null;var a=t.scale,o=t.duplicateDomain,i=t.type,s=t.range,l=t.realScaleType==="scaleBand"?a.bandwidth()/2:2,u=(n||r)&&i==="category"&&a.bandwidth?a.bandwidth()/l:0;if(u=t.axisType==="angleAxis"&&(s==null?void 0:s.length)>=2?In(s[0]-s[1])*2*u:u,n&&(t.ticks||t.niceTicks)){var p=(t.ticks||t.niceTicks).map(function(c){var f=o?o.indexOf(c):c;return{coordinate:a(f)+u,value:c,offset:u}});return p.filter(function(c){return!vs(c.coordinate)})}return t.isCategorical&&t.categoricalDomain?t.categoricalDomain.map(function(c,f){return{coordinate:a(c)+u,value:c,index:f,offset:u}}):a.ticks&&!r?a.ticks(t.tickCount).map(function(c){return{coordinate:a(c)+u,value:c,offset:u}}):a.domain().map(function(c,f){return{coordinate:a(c)+u,value:o?o[c]:c,index:f,offset:u}})},Im=new WeakMap,pc=function(t,n){if(typeof n!="function")return t;Im.has(t)||Im.set(t,new WeakMap);var r=Im.get(t);if(r.has(n))return r.get(n);var a=function(){t.apply(void 0,arguments),n.apply(void 0,arguments)};return r.set(n,a),a},BX=function(t,n,r){var a=t.scale,o=t.type,i=t.layout,s=t.axisType;if(a==="auto")return i==="radial"&&s==="radiusAxis"?{scale:Ll(),realScaleType:"band"}:i==="radial"&&s==="angleAxis"?{scale:Dp(),realScaleType:"linear"}:o==="category"&&n&&(n.indexOf("LineChart")>=0||n.indexOf("AreaChart")>=0||n.indexOf("ComposedChart")>=0&&!r)?{scale:sl(),realScaleType:"point"}:o==="category"?{scale:Ll(),realScaleType:"band"}:{scale:Dp(),realScaleType:"linear"};if(xo(a)){var l="scale".concat(nd(a));return{scale:(B3[l]||sl)(),realScaleType:B3[l]?l:"point"}}return ae(a)?{scale:a}:{scale:sl(),realScaleType:"point"}},K3=1e-4,zX=function(t){var n=t.domain();if(!(!n||n.length<=2)){var r=n.length,a=t.range(),o=Math.min(a[0],a[1])-K3,i=Math.max(a[0],a[1])+K3,s=t(n[0]),l=t(n[r-1]);(s<o||s>i||l<o||l>i)&&t.domain([n[0],n[r-1]])}},HX=function(t,n){if(!t)return null;for(var r=0,a=t.length;r<a;r++)if(t[r].item===n)return t[r].position;return null},GX=function(t,n){if(!n||n.length!==2||!V(n[0])||!V(n[1]))return t;var r=Math.min(n[0],n[1]),a=Math.max(n[0],n[1]),o=[t[0],t[1]];return(!V(t[0])||t[0]<r)&&(o[0]=r),(!V(t[1])||t[1]>a)&&(o[1]=a),o[0]>a&&(o[0]=a),o[1]<r&&(o[1]=r),o},UX=function(t){var n=t.length;if(!(n<=0))for(var r=0,a=t[0].length;r<a;++r)for(var o=0,i=0,s=0;s<n;++s){var l=vs(t[s][r][1])?t[s][r][0]:t[s][r][1];l>=0?(t[s][r][0]=o,t[s][r][1]=o+l,o=t[s][r][1]):(t[s][r][0]=i,t[s][r][1]=i+l,i=t[s][r][1])}},WX=function(t){var n=t.length;if(!(n<=0))for(var r=0,a=t[0].length;r<a;++r)for(var o=0,i=0;i<n;++i){var s=vs(t[i][r][1])?t[i][r][0]:t[i][r][1];s>=0?(t[i][r][0]=o,t[i][r][1]=o+s,o=t[i][r][1]):(t[i][r][0]=0,t[i][r][1]=0)}},qX={sign:UX,expand:iI,none:ji,silhouette:sI,wiggle:lI,positive:WX},VX=function(t,n,r){var a=n.map(function(s){return s.props.dataKey}),o=qX[r],i=oI().keys(a).value(function(s,l){return+It(s,l,0)}).order(i0).offset(o);return i(t)},KX=function(t,n,r,a,o,i){if(!t)return null;var s=i?n.reverse():n,l={},u=s.reduce(function(c,f){var m,h=(m=f.type)!==null&&m!==void 0&&m.defaultProps?Ve(Ve({},f.type.defaultProps),f.props):f.props,d=h.stackId,g=h.hide;if(g)return c;var v=h[r],y=c[v]||{hasStack:!1,stackGroups:{}};if(ot(d)){var x=y.stackGroups[d]||{numericAxisId:r,cateAxisId:a,items:[]};x.items.push(f),y.hasStack=!0,y.stackGroups[d]=x}else y.stackGroups[ys("_stackId_")]={numericAxisId:r,cateAxisId:a,items:[f]};return Ve(Ve({},c),{},fi({},v,y))},l),p={};return Object.keys(u).reduce(function(c,f){var m=u[f];if(m.hasStack){var h={};m.stackGroups=Object.keys(m.stackGroups).reduce(function(d,g){var v=m.stackGroups[g];return Ve(Ve({},d),{},fi({},g,{numericAxisId:r,cateAxisId:a,items:v.items,stackedData:VX(t,v.items,o)}))},h)}return Ve(Ve({},c),{},fi({},f,m))},p)},XX=function(t,n){var r=n.realScaleType,a=n.type,o=n.tickCount,i=n.originalDomain,s=n.allowDecimals,l=r||n.scale;if(l!=="auto"&&l!=="linear")return null;if(o&&a==="number"&&i&&(i[0]==="auto"||i[1]==="auto")){var u=t.domain();if(!u.length)return null;var p=iX(u,o,s);return t.domain([hd(p),pa(p)]),{niceTicks:p}}if(o&&a==="number"){var c=t.domain(),f=sX(c,o,s);return{niceTicks:f}}return null};function Up(e){var t=e.axis,n=e.ticks,r=e.bandSize,a=e.entry,o=e.index,i=e.dataKey;if(t.type==="category"){if(!t.allowDuplicatedCategory&&t.dataKey&&!le(a[t.dataKey])){var s=mp(n,"value",a[t.dataKey]);if(s)return s.coordinate+r/2}return n[o]?n[o].coordinate+r/2:null}var l=It(a,le(i)?t.dataKey:i);return le(l)?null:t.scale(l)}var X3=function(t){var n=t.axis,r=t.ticks,a=t.offset,o=t.bandSize,i=t.entry,s=t.index;if(n.type==="category")return r[s]?r[s].coordinate+a:null;var l=It(i,n.dataKey,n.domain[s]);return le(l)?null:n.scale(l)-o/2+a},YX=function(t){var n=t.numericAxis,r=n.scale.domain();if(n.type==="number"){var a=Math.min(r[0],r[1]),o=Math.max(r[0],r[1]);return a<=0&&o>=0?0:o<0?o:a}return r[0]},QX=function(t,n){var r,a=(r=t.type)!==null&&r!==void 0&&r.defaultProps?Ve(Ve({},t.type.defaultProps),t.props):t.props,o=a.stackId;if(ot(o)){var i=n[o];if(i){var s=i.items.indexOf(t);return s>=0?i.stackedData[s]:null}}return null},ZX=function(t){return t.reduce(function(n,r){return[hd(r.concat([n[0]]).filter(V)),pa(r.concat([n[1]]).filter(V))]},[1/0,-1/0])},dP=function(t,n,r){return Object.keys(t).reduce(function(a,o){var i=t[o],s=i.stackedData,l=s.reduce(function(u,p){var c=ZX(p.slice(n,r+1));return[Math.min(u[0],c[0]),Math.max(u[1],c[1])]},[1/0,-1/0]);return[Math.min(l[0],a[0]),Math.max(l[1],a[1])]},[1/0,-1/0]).map(function(a){return a===1/0||a===-1/0?0:a})},Y3=/^dataMin[\s]*-[\s]*([0-9]+([.]{1}[0-9]+){0,1})$/,Q3=/^dataMax[\s]*\+[\s]*([0-9]+([.]{1}[0-9]+){0,1})$/,I0=function(t,n,r){if(ae(t))return t(n,r);if(!Array.isArray(t))return n;var a=[];if(V(t[0]))a[0]=r?t[0]:Math.min(t[0],n[0]);else if(Y3.test(t[0])){var o=+Y3.exec(t[0])[1];a[0]=n[0]-o}else ae(t[0])?a[0]=t[0](n[0]):a[0]=n[0];if(V(t[1]))a[1]=r?t[1]:Math.max(t[1],n[1]);else if(Q3.test(t[1])){var i=+Q3.exec(t[1])[1];a[1]=n[1]+i}else ae(t[1])?a[1]=t[1](n[1]):a[1]=n[1];return a},Wp=function(t,n,r){if(t&&t.scale&&t.scale.bandwidth){var a=t.scale.bandwidth();if(!r||a>0)return a}if(t&&n&&n.length>=2){for(var o=tg(n,function(c){return c.coordinate}),i=1/0,s=1,l=o.length;s<l;s++){var u=o[s],p=o[s-1];i=Math.min((u.coordinate||0)-(p.coordinate||0),i)}return i===1/0?0:i}return r?void 0:0},Z3=function(t,n,r){return!t||!t.length||Fi(t,vn(r,"type.defaultProps.domain"))?n:t},mP=function(t,n){var r=t.type.defaultProps?Ve(Ve({},t.type.defaultProps),t.props):t.props,a=r.dataKey,o=r.name,i=r.unit,s=r.formatter,l=r.tooltipType,u=r.chartType,p=r.hide;return Ve(Ve({},ie(t,!1)),{},{dataKey:a,unit:i,formatter:s,name:o||a,color:_g(t),value:It(n,a),type:l,payload:n,chartType:u,hide:p})};function Vl(e){"@babel/helpers - typeof";return Vl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Vl(e)}function J3(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function eb(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?J3(Object(n),!0).forEach(function(r){JX(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):J3(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function JX(e,t,n){return t=eY(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function eY(e){var t=tY(e,"string");return Vl(t)=="symbol"?t:t+""}function tY(e,t){if(Vl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Vl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var qp=Math.PI/180,nY=function(t){return t*180/Math.PI},mt=function(t,n,r,a){return{x:t+Math.cos(-qp*a)*r,y:n+Math.sin(-qp*a)*r}},rY=function(t,n){var r=t.x,a=t.y,o=n.x,i=n.y;return Math.sqrt(Math.pow(r-o,2)+Math.pow(a-i,2))},aY=function(t,n){var r=t.x,a=t.y,o=n.cx,i=n.cy,s=rY({x:r,y:a},{x:o,y:i});if(s<=0)return{radius:s};var l=(r-o)/s,u=Math.acos(l);return a>i&&(u=2*Math.PI-u),{radius:s,angle:nY(u),angleInRadian:u}},oY=function(t){var n=t.startAngle,r=t.endAngle,a=Math.floor(n/360),o=Math.floor(r/360),i=Math.min(a,o);return{startAngle:n-i*360,endAngle:r-i*360}},iY=function(t,n){var r=n.startAngle,a=n.endAngle,o=Math.floor(r/360),i=Math.floor(a/360),s=Math.min(o,i);return t+s*360},tb=function(t,n){var r=t.x,a=t.y,o=aY({x:r,y:a},n),i=o.radius,s=o.angle,l=n.innerRadius,u=n.outerRadius;if(i<l||i>u)return!1;if(i===0)return!0;var p=oY(n),c=p.startAngle,f=p.endAngle,m=s,h;if(c<=f){for(;m>f;)m-=360;for(;m<c;)m+=360;h=m>=c&&m<=f}else{for(;m>c;)m-=360;for(;m<f;)m+=360;h=m>=f&&m<=c}return h?eb(eb({},n),{},{radius:i,angle:iY(m,n)}):null};function Kl(e){"@babel/helpers - typeof";return Kl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Kl(e)}var sY=["offset"];function lY(e){return fY(e)||pY(e)||cY(e)||uY()}function uY(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function cY(e,t){if(e){if(typeof e=="string")return D0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return D0(e,t)}}function pY(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function fY(e){if(Array.isArray(e))return D0(e)}function D0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function dY(e,t){if(e==null)return{};var n=mY(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function mY(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function nb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function nt(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?nb(Object(n),!0).forEach(function(r){hY(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):nb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function hY(e,t,n){return t=vY(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function vY(e){var t=yY(e,"string");return Kl(t)=="symbol"?t:t+""}function yY(e,t){if(Kl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Kl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function Xl(){return Xl=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Xl.apply(this,arguments)}var gY=function(t){var n=t.value,r=t.formatter,a=le(t.children)?n:t.children;return ae(r)?r(a):a},xY=function(t,n){var r=In(n-t),a=Math.min(Math.abs(n-t),360);return r*a},wY=function(t,n,r){var a=t.position,o=t.viewBox,i=t.offset,s=t.className,l=o,u=l.cx,p=l.cy,c=l.innerRadius,f=l.outerRadius,m=l.startAngle,h=l.endAngle,d=l.clockWise,g=(c+f)/2,v=xY(m,h),y=v>=0?1:-1,x,P;a==="insideStart"?(x=m+y*i,P=d):a==="insideEnd"?(x=h-y*i,P=!d):a==="end"&&(x=h+y*i,P=d),P=v<=0?P:!P;var w=mt(u,p,g,x),S=mt(u,p,g,x+(P?1:-1)*359),O="M".concat(w.x,",").concat(w.y,`
    A`).concat(g,",").concat(g,",0,1,").concat(P?0:1,`,
    `).concat(S.x,",").concat(S.y),C=le(t.id)?ys("recharts-radial-line-"):t.id;return E.createElement("text",Xl({},r,{dominantBaseline:"central",className:ue("recharts-radial-bar-label",s)}),E.createElement("defs",null,E.createElement("path",{id:C,d:O})),E.createElement("textPath",{xlinkHref:"#".concat(C)},n))},bY=function(t){var n=t.viewBox,r=t.offset,a=t.position,o=n,i=o.cx,s=o.cy,l=o.innerRadius,u=o.outerRadius,p=o.startAngle,c=o.endAngle,f=(p+c)/2;if(a==="outside"){var m=mt(i,s,u+r,f),h=m.x,d=m.y;return{x:h,y:d,textAnchor:h>=i?"start":"end",verticalAnchor:"middle"}}if(a==="center")return{x:i,y:s,textAnchor:"middle",verticalAnchor:"middle"};if(a==="centerTop")return{x:i,y:s,textAnchor:"middle",verticalAnchor:"start"};if(a==="centerBottom")return{x:i,y:s,textAnchor:"middle",verticalAnchor:"end"};var g=(l+u)/2,v=mt(i,s,g,f),y=v.x,x=v.y;return{x:y,y:x,textAnchor:"middle",verticalAnchor:"middle"}},PY=function(t){var n=t.viewBox,r=t.parentViewBox,a=t.offset,o=t.position,i=n,s=i.x,l=i.y,u=i.width,p=i.height,c=p>=0?1:-1,f=c*a,m=c>0?"end":"start",h=c>0?"start":"end",d=u>=0?1:-1,g=d*a,v=d>0?"end":"start",y=d>0?"start":"end";if(o==="top"){var x={x:s+u/2,y:l-c*a,textAnchor:"middle",verticalAnchor:m};return nt(nt({},x),r?{height:Math.max(l-r.y,0),width:u}:{})}if(o==="bottom"){var P={x:s+u/2,y:l+p+f,textAnchor:"middle",verticalAnchor:h};return nt(nt({},P),r?{height:Math.max(r.y+r.height-(l+p),0),width:u}:{})}if(o==="left"){var w={x:s-g,y:l+p/2,textAnchor:v,verticalAnchor:"middle"};return nt(nt({},w),r?{width:Math.max(w.x-r.x,0),height:p}:{})}if(o==="right"){var S={x:s+u+g,y:l+p/2,textAnchor:y,verticalAnchor:"middle"};return nt(nt({},S),r?{width:Math.max(r.x+r.width-S.x,0),height:p}:{})}var O=r?{width:u,height:p}:{};return o==="insideLeft"?nt({x:s+g,y:l+p/2,textAnchor:y,verticalAnchor:"middle"},O):o==="insideRight"?nt({x:s+u-g,y:l+p/2,textAnchor:v,verticalAnchor:"middle"},O):o==="insideTop"?nt({x:s+u/2,y:l+f,textAnchor:"middle",verticalAnchor:h},O):o==="insideBottom"?nt({x:s+u/2,y:l+p-f,textAnchor:"middle",verticalAnchor:m},O):o==="insideTopLeft"?nt({x:s+g,y:l+f,textAnchor:y,verticalAnchor:h},O):o==="insideTopRight"?nt({x:s+u-g,y:l+f,textAnchor:v,verticalAnchor:h},O):o==="insideBottomLeft"?nt({x:s+g,y:l+p-f,textAnchor:y,verticalAnchor:m},O):o==="insideBottomRight"?nt({x:s+u-g,y:l+p-f,textAnchor:v,verticalAnchor:m},O):fs(o)&&(V(o.x)||Xa(o.x))&&(V(o.y)||Xa(o.y))?nt({x:s+wo(o.x,u),y:l+wo(o.y,p),textAnchor:"end",verticalAnchor:"end"},O):nt({x:s+u/2,y:l+p/2,textAnchor:"middle",verticalAnchor:"middle"},O)},SY=function(t){return"cx"in t&&V(t.cx)};function kt(e){var t=e.offset,n=t===void 0?5:t,r=dY(e,sY),a=nt({offset:n},r),o=a.viewBox,i=a.position,s=a.value,l=a.children,u=a.content,p=a.className,c=p===void 0?"":p,f=a.textBreakAll;if(!o||le(s)&&le(l)&&!k.isValidElement(u)&&!ae(u))return null;if(k.isValidElement(u))return k.cloneElement(u,a);var m;if(ae(u)){if(m=k.createElement(u,a),k.isValidElement(m))return m}else m=gY(a);var h=SY(o),d=ie(a,!0);if(h&&(i==="insideStart"||i==="insideEnd"||i==="end"))return wY(a,m,d);var g=h?bY(a):PY(a);return E.createElement(Tp,Xl({className:ue("recharts-label",c)},d,g,{breakAll:f}),m)}kt.displayName="Label";var hP=function(t){var n=t.cx,r=t.cy,a=t.angle,o=t.startAngle,i=t.endAngle,s=t.r,l=t.radius,u=t.innerRadius,p=t.outerRadius,c=t.x,f=t.y,m=t.top,h=t.left,d=t.width,g=t.height,v=t.clockWise,y=t.labelViewBox;if(y)return y;if(V(d)&&V(g)){if(V(c)&&V(f))return{x:c,y:f,width:d,height:g};if(V(m)&&V(h))return{x:m,y:h,width:d,height:g}}return V(c)&&V(f)?{x:c,y:f,width:0,height:0}:V(n)&&V(r)?{cx:n,cy:r,startAngle:o||a||0,endAngle:i||a||0,innerRadius:u||0,outerRadius:p||l||s||0,clockWise:v}:t.viewBox?t.viewBox:{}},OY=function(t,n){return t?t===!0?E.createElement(kt,{key:"label-implicit",viewBox:n}):ot(t)?E.createElement(kt,{key:"label-implicit",viewBox:n,value:t}):k.isValidElement(t)?t.type===kt?k.cloneElement(t,{key:"label-implicit",viewBox:n}):E.createElement(kt,{key:"label-implicit",content:t,viewBox:n}):ae(t)?E.createElement(kt,{key:"label-implicit",content:t,viewBox:n}):fs(t)?E.createElement(kt,Xl({viewBox:n},t,{key:"label-implicit"})):null:null},kY=function(t,n){var r=arguments.length>2&&arguments[2]!==void 0?arguments[2]:!0;if(!t||!t.children&&r&&!t.label)return null;var a=t.children,o=hP(t),i=yn(a,kt).map(function(l,u){return k.cloneElement(l,{viewBox:n||o,key:"label-".concat(u)})});if(!r)return i;var s=OY(t.label,n||o);return[s].concat(lY(i))};kt.parseViewBox=hP;kt.renderCallByParent=kY;function CY(e){var t=e==null?0:e.length;return t?e[t-1]:void 0}var _Y=CY;const AY=_e(_Y);function Yl(e){"@babel/helpers - typeof";return Yl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Yl(e)}var EY=["valueAccessor"],TY=["data","dataKey","clockWise","id","textBreakAll"];function jY(e){return RY(e)||MY(e)||$Y(e)||NY()}function NY(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function $Y(e,t){if(e){if(typeof e=="string")return L0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return L0(e,t)}}function MY(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function RY(e){if(Array.isArray(e))return L0(e)}function L0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function Vp(){return Vp=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Vp.apply(this,arguments)}function rb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function ab(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?rb(Object(n),!0).forEach(function(r){IY(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):rb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function IY(e,t,n){return t=DY(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function DY(e){var t=LY(e,"string");return Yl(t)=="symbol"?t:t+""}function LY(e,t){if(Yl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Yl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function ob(e,t){if(e==null)return{};var n=FY(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function FY(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}var BY=function(t){return Array.isArray(t.value)?AY(t.value):t.value};function Nr(e){var t=e.valueAccessor,n=t===void 0?BY:t,r=ob(e,EY),a=r.data,o=r.dataKey,i=r.clockWise,s=r.id,l=r.textBreakAll,u=ob(r,TY);return!a||!a.length?null:E.createElement(Ne,{className:"recharts-label-list"},a.map(function(p,c){var f=le(o)?n(p,c):It(p&&p.payload,o),m=le(s)?{}:{id:"".concat(s,"-").concat(c)};return E.createElement(kt,Vp({},ie(p,!0),u,m,{parentViewBox:p.parentViewBox,value:f,textBreakAll:l,viewBox:kt.parseViewBox(le(i)?p:ab(ab({},p),{},{clockWise:i})),key:"label-".concat(c),index:c}))}))}Nr.displayName="LabelList";function zY(e,t){return e?e===!0?E.createElement(Nr,{key:"labelList-implicit",data:t}):E.isValidElement(e)||ae(e)?E.createElement(Nr,{key:"labelList-implicit",data:t,content:e}):fs(e)?E.createElement(Nr,Vp({data:t},e,{key:"labelList-implicit"})):null:null}function HY(e,t){var n=arguments.length>2&&arguments[2]!==void 0?arguments[2]:!0;if(!e||!e.children&&n&&!e.label)return null;var r=e.children,a=yn(r,Nr).map(function(i,s){return k.cloneElement(i,{data:t,key:"labelList-".concat(s)})});if(!n)return a;var o=zY(e.label,t);return[o].concat(jY(a))}Nr.renderCallByParent=HY;function Ql(e){"@babel/helpers - typeof";return Ql=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Ql(e)}function F0(){return F0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},F0.apply(this,arguments)}function ib(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function sb(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?ib(Object(n),!0).forEach(function(r){GY(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):ib(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function GY(e,t,n){return t=UY(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function UY(e){var t=WY(e,"string");return Ql(t)=="symbol"?t:t+""}function WY(e,t){if(Ql(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Ql(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var qY=function(t,n){var r=In(n-t),a=Math.min(Math.abs(n-t),359.999);return r*a},fc=function(t){var n=t.cx,r=t.cy,a=t.radius,o=t.angle,i=t.sign,s=t.isExternal,l=t.cornerRadius,u=t.cornerIsExternal,p=l*(s?1:-1)+a,c=Math.asin(l/p)/qp,f=u?o:o+i*c,m=mt(n,r,p,f),h=mt(n,r,a,f),d=u?o-i*c:o,g=mt(n,r,p*Math.cos(c*qp),d);return{center:m,circleTangency:h,lineTangency:g,theta:c}},vP=function(t){var n=t.cx,r=t.cy,a=t.innerRadius,o=t.outerRadius,i=t.startAngle,s=t.endAngle,l=qY(i,s),u=i+l,p=mt(n,r,o,i),c=mt(n,r,o,u),f="M ".concat(p.x,",").concat(p.y,`
    A `).concat(o,",").concat(o,`,0,
    `).concat(+(Math.abs(l)>180),",").concat(+(i>u),`,
    `).concat(c.x,",").concat(c.y,`
  `);if(a>0){var m=mt(n,r,a,i),h=mt(n,r,a,u);f+="L ".concat(h.x,",").concat(h.y,`
            A `).concat(a,",").concat(a,`,0,
            `).concat(+(Math.abs(l)>180),",").concat(+(i<=u),`,
            `).concat(m.x,",").concat(m.y," Z")}else f+="L ".concat(n,",").concat(r," Z");return f},VY=function(t){var n=t.cx,r=t.cy,a=t.innerRadius,o=t.outerRadius,i=t.cornerRadius,s=t.forceCornerRadius,l=t.cornerIsExternal,u=t.startAngle,p=t.endAngle,c=In(p-u),f=fc({cx:n,cy:r,radius:o,angle:u,sign:c,cornerRadius:i,cornerIsExternal:l}),m=f.circleTangency,h=f.lineTangency,d=f.theta,g=fc({cx:n,cy:r,radius:o,angle:p,sign:-c,cornerRadius:i,cornerIsExternal:l}),v=g.circleTangency,y=g.lineTangency,x=g.theta,P=l?Math.abs(u-p):Math.abs(u-p)-d-x;if(P<0)return s?"M ".concat(h.x,",").concat(h.y,`
        a`).concat(i,",").concat(i,",0,0,1,").concat(i*2,`,0
        a`).concat(i,",").concat(i,",0,0,1,").concat(-i*2,`,0
      `):vP({cx:n,cy:r,innerRadius:a,outerRadius:o,startAngle:u,endAngle:p});var w="M ".concat(h.x,",").concat(h.y,`
    A`).concat(i,",").concat(i,",0,0,").concat(+(c<0),",").concat(m.x,",").concat(m.y,`
    A`).concat(o,",").concat(o,",0,").concat(+(P>180),",").concat(+(c<0),",").concat(v.x,",").concat(v.y,`
    A`).concat(i,",").concat(i,",0,0,").concat(+(c<0),",").concat(y.x,",").concat(y.y,`
  `);if(a>0){var S=fc({cx:n,cy:r,radius:a,angle:u,sign:c,isExternal:!0,cornerRadius:i,cornerIsExternal:l}),O=S.circleTangency,C=S.lineTangency,_=S.theta,T=fc({cx:n,cy:r,radius:a,angle:p,sign:-c,isExternal:!0,cornerRadius:i,cornerIsExternal:l}),A=T.circleTangency,j=T.lineTangency,$=T.theta,M=l?Math.abs(u-p):Math.abs(u-p)-_-$;if(M<0&&i===0)return"".concat(w,"L").concat(n,",").concat(r,"Z");w+="L".concat(j.x,",").concat(j.y,`
      A`).concat(i,",").concat(i,",0,0,").concat(+(c<0),",").concat(A.x,",").concat(A.y,`
      A`).concat(a,",").concat(a,",0,").concat(+(M>180),",").concat(+(c>0),",").concat(O.x,",").concat(O.y,`
      A`).concat(i,",").concat(i,",0,0,").concat(+(c<0),",").concat(C.x,",").concat(C.y,"Z")}else w+="L".concat(n,",").concat(r,"Z");return w},KY={cx:0,cy:0,innerRadius:0,outerRadius:0,startAngle:0,endAngle:0,cornerRadius:0,forceCornerRadius:!1,cornerIsExternal:!1},yP=function(t){var n=sb(sb({},KY),t),r=n.cx,a=n.cy,o=n.innerRadius,i=n.outerRadius,s=n.cornerRadius,l=n.forceCornerRadius,u=n.cornerIsExternal,p=n.startAngle,c=n.endAngle,f=n.className;if(i<o||p===c)return null;var m=ue("recharts-sector",f),h=i-o,d=wo(s,h,0,!0),g;return d>0&&Math.abs(p-c)<360?g=VY({cx:r,cy:a,innerRadius:o,outerRadius:i,cornerRadius:Math.min(d,h/2),forceCornerRadius:l,cornerIsExternal:u,startAngle:p,endAngle:c}):g=vP({cx:r,cy:a,innerRadius:o,outerRadius:i,startAngle:p,endAngle:c}),E.createElement("path",F0({},ie(n,!0),{className:m,d:g,role:"img"}))};function Zl(e){"@babel/helpers - typeof";return Zl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Zl(e)}function B0(){return B0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},B0.apply(this,arguments)}function lb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function ub(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?lb(Object(n),!0).forEach(function(r){XY(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):lb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function XY(e,t,n){return t=YY(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function YY(e){var t=QY(e,"string");return Zl(t)=="symbol"?t:t+""}function QY(e,t){if(Zl(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Zl(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var cb={curveBasisClosed:KR,curveBasisOpen:XR,curveBasis:VR,curveBumpX:$R,curveBumpY:MR,curveLinearClosed:YR,curveLinear:ad,curveMonotoneX:QR,curveMonotoneY:ZR,curveNatural:JR,curveStep:eI,curveStepAfter:nI,curveStepBefore:tI},dc=function(t){return t.x===+t.x&&t.y===+t.y},zs=function(t){return t.x},Hs=function(t){return t.y},ZY=function(t,n){if(ae(t))return t;var r="curve".concat(nd(t));return(r==="curveMonotone"||r==="curveBump")&&n?cb["".concat(r).concat(n==="vertical"?"Y":"X")]:cb[r]||ad},JY=function(t){var n=t.type,r=n===void 0?"linear":n,a=t.points,o=a===void 0?[]:a,i=t.baseLine,s=t.layout,l=t.connectNulls,u=l===void 0?!1:l,p=ZY(r,s),c=u?o.filter(function(d){return dc(d)}):o,f;if(Array.isArray(i)){var m=u?i.filter(function(d){return dc(d)}):i,h=c.map(function(d,g){return ub(ub({},d),{},{base:m[g]})});return s==="vertical"?f=rc().y(Hs).x1(zs).x0(function(d){return d.base.x}):f=rc().x(zs).y1(Hs).y0(function(d){return d.base.y}),f.defined(dc).curve(p),f(h)}return s==="vertical"&&V(i)?f=rc().y(Hs).x1(zs).x0(i):V(i)?f=rc().x(zs).y1(Hs).y0(i):f=y8().x(zs).y(Hs),f.defined(dc).curve(p),f(c)},di=function(t){var n=t.className,r=t.points,a=t.path,o=t.pathRef;if((!r||!r.length)&&!a)return null;var i=r&&r.length?JY(t):a;return k.createElement("path",B0({},ie(t,!1),hp(t),{className:ue("recharts-curve",n),d:i,ref:o}))},gP={exports:{}},eQ="SECRET_DO_NOT_PASS_THIS_OR_YOU_WILL_BE_FIRED",tQ=eQ,nQ=tQ;function xP(){}function wP(){}wP.resetWarningCache=xP;var rQ=function(){function e(r,a,o,i,s,l){if(l!==nQ){var u=new Error("Calling PropTypes validators directly is not supported by the `prop-types` package. Use PropTypes.checkPropTypes() to call them. Read more at http://fb.me/use-check-prop-types");throw u.name="Invariant Violation",u}}e.isRequired=e;function t(){return e}var n={array:e,bigint:e,bool:e,func:e,number:e,object:e,string:e,symbol:e,any:e,arrayOf:t,element:e,elementType:e,instanceOf:t,node:e,objectOf:t,oneOf:t,oneOfType:t,shape:t,exact:t,checkPropTypes:wP,resetWarningCache:xP};return n.PropTypes=n,n};gP.exports=rQ();var aQ=gP.exports;const ye=_e(aQ);var oQ=Object.getOwnPropertyNames,iQ=Object.getOwnPropertySymbols,sQ=Object.prototype.hasOwnProperty;function pb(e,t){return function(r,a,o){return e(r,a,o)&&t(r,a,o)}}function mc(e){return function(n,r,a){if(!n||!r||typeof n!="object"||typeof r!="object")return e(n,r,a);var o=a.cache,i=o.get(n),s=o.get(r);if(i&&s)return i===r&&s===n;o.set(n,r),o.set(r,n);var l=e(n,r,a);return o.delete(n),o.delete(r),l}}function fb(e){return oQ(e).concat(iQ(e))}var lQ=Object.hasOwn||function(e,t){return sQ.call(e,t)};function To(e,t){return e===t||!e&&!t&&e!==e&&t!==t}var uQ="__v",cQ="__o",pQ="_owner",db=Object.getOwnPropertyDescriptor,mb=Object.keys;function fQ(e,t,n){var r=e.length;if(t.length!==r)return!1;for(;r-- >0;)if(!n.equals(e[r],t[r],r,r,e,t,n))return!1;return!0}function dQ(e,t){return To(e.getTime(),t.getTime())}function mQ(e,t){return e.name===t.name&&e.message===t.message&&e.cause===t.cause&&e.stack===t.stack}function hQ(e,t){return e===t}function hb(e,t,n){var r=e.size;if(r!==t.size)return!1;if(!r)return!0;for(var a=new Array(r),o=e.entries(),i,s,l=0;(i=o.next())&&!i.done;){for(var u=t.entries(),p=!1,c=0;(s=u.next())&&!s.done;){if(a[c]){c++;continue}var f=i.value,m=s.value;if(n.equals(f[0],m[0],l,c,e,t,n)&&n.equals(f[1],m[1],f[0],m[0],e,t,n)){p=a[c]=!0;break}c++}if(!p)return!1;l++}return!0}var vQ=To;function yQ(e,t,n){var r=mb(e),a=r.length;if(mb(t).length!==a)return!1;for(;a-- >0;)if(!bP(e,t,n,r[a]))return!1;return!0}function Gs(e,t,n){var r=fb(e),a=r.length;if(fb(t).length!==a)return!1;for(var o,i,s;a-- >0;)if(o=r[a],!bP(e,t,n,o)||(i=db(e,o),s=db(t,o),(i||s)&&(!i||!s||i.configurable!==s.configurable||i.enumerable!==s.enumerable||i.writable!==s.writable)))return!1;return!0}function gQ(e,t){return To(e.valueOf(),t.valueOf())}function xQ(e,t){return e.source===t.source&&e.flags===t.flags}function vb(e,t,n){var r=e.size;if(r!==t.size)return!1;if(!r)return!0;for(var a=new Array(r),o=e.values(),i,s;(i=o.next())&&!i.done;){for(var l=t.values(),u=!1,p=0;(s=l.next())&&!s.done;){if(!a[p]&&n.equals(i.value,s.value,i.value,s.value,e,t,n)){u=a[p]=!0;break}p++}if(!u)return!1}return!0}function wQ(e,t){var n=e.length;if(t.length!==n)return!1;for(;n-- >0;)if(e[n]!==t[n])return!1;return!0}function bQ(e,t){return e.hostname===t.hostname&&e.pathname===t.pathname&&e.protocol===t.protocol&&e.port===t.port&&e.hash===t.hash&&e.username===t.username&&e.password===t.password}function bP(e,t,n,r){return(r===pQ||r===cQ||r===uQ)&&(e.$$typeof||t.$$typeof)?!0:lQ(t,r)&&n.equals(e[r],t[r],r,r,e,t,n)}var PQ="[object Arguments]",SQ="[object Boolean]",OQ="[object Date]",kQ="[object Error]",CQ="[object Map]",_Q="[object Number]",AQ="[object Object]",EQ="[object RegExp]",TQ="[object Set]",jQ="[object String]",NQ="[object URL]",$Q=Array.isArray,yb=typeof ArrayBuffer=="function"&&ArrayBuffer.isView?ArrayBuffer.isView:null,gb=Object.assign,MQ=Object.prototype.toString.call.bind(Object.prototype.toString);function RQ(e){var t=e.areArraysEqual,n=e.areDatesEqual,r=e.areErrorsEqual,a=e.areFunctionsEqual,o=e.areMapsEqual,i=e.areNumbersEqual,s=e.areObjectsEqual,l=e.arePrimitiveWrappersEqual,u=e.areRegExpsEqual,p=e.areSetsEqual,c=e.areTypedArraysEqual,f=e.areUrlsEqual;return function(h,d,g){if(h===d)return!0;if(h==null||d==null)return!1;var v=typeof h;if(v!==typeof d)return!1;if(v!=="object")return v==="number"?i(h,d,g):v==="function"?a(h,d,g):!1;var y=h.constructor;if(y!==d.constructor)return!1;if(y===Object)return s(h,d,g);if($Q(h))return t(h,d,g);if(yb!=null&&yb(h))return c(h,d,g);if(y===Date)return n(h,d,g);if(y===RegExp)return u(h,d,g);if(y===Map)return o(h,d,g);if(y===Set)return p(h,d,g);var x=MQ(h);return x===OQ?n(h,d,g):x===EQ?u(h,d,g):x===CQ?o(h,d,g):x===TQ?p(h,d,g):x===AQ?typeof h.then!="function"&&typeof d.then!="function"&&s(h,d,g):x===NQ?f(h,d,g):x===kQ?r(h,d,g):x===PQ?s(h,d,g):x===SQ||x===_Q||x===jQ?l(h,d,g):!1}}function IQ(e){var t=e.circular,n=e.createCustomConfig,r=e.strict,a={areArraysEqual:r?Gs:fQ,areDatesEqual:dQ,areErrorsEqual:mQ,areFunctionsEqual:hQ,areMapsEqual:r?pb(hb,Gs):hb,areNumbersEqual:vQ,areObjectsEqual:r?Gs:yQ,arePrimitiveWrappersEqual:gQ,areRegExpsEqual:xQ,areSetsEqual:r?pb(vb,Gs):vb,areTypedArraysEqual:r?Gs:wQ,areUrlsEqual:bQ};if(n&&(a=gb({},a,n(a))),t){var o=mc(a.areArraysEqual),i=mc(a.areMapsEqual),s=mc(a.areObjectsEqual),l=mc(a.areSetsEqual);a=gb({},a,{areArraysEqual:o,areMapsEqual:i,areObjectsEqual:s,areSetsEqual:l})}return a}function DQ(e){return function(t,n,r,a,o,i,s){return e(t,n,s)}}function LQ(e){var t=e.circular,n=e.comparator,r=e.createState,a=e.equals,o=e.strict;if(r)return function(l,u){var p=r(),c=p.cache,f=c===void 0?t?new WeakMap:void 0:c,m=p.meta;return n(l,u,{cache:f,equals:a,meta:m,strict:o})};if(t)return function(l,u){return n(l,u,{cache:new WeakMap,equals:a,meta:void 0,strict:o})};var i={cache:void 0,equals:a,meta:void 0,strict:o};return function(l,u){return n(l,u,i)}}var FQ=Ia();Ia({strict:!0});Ia({circular:!0});Ia({circular:!0,strict:!0});Ia({createInternalComparator:function(){return To}});Ia({strict:!0,createInternalComparator:function(){return To}});Ia({circular:!0,createInternalComparator:function(){return To}});Ia({circular:!0,createInternalComparator:function(){return To},strict:!0});function Ia(e){e===void 0&&(e={});var t=e.circular,n=t===void 0?!1:t,r=e.createInternalComparator,a=e.createState,o=e.strict,i=o===void 0?!1:o,s=IQ(e),l=RQ(s),u=r?r(l):DQ(l);return LQ({circular:n,comparator:l,createState:a,equals:u,strict:i})}function BQ(e){typeof requestAnimationFrame<"u"&&requestAnimationFrame(e)}function xb(e){var t=arguments.length>1&&arguments[1]!==void 0?arguments[1]:0,n=-1,r=function a(o){n<0&&(n=o),o-n>t?(e(o),n=-1):BQ(a)};requestAnimationFrame(r)}function z0(e){"@babel/helpers - typeof";return z0=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},z0(e)}function zQ(e){return WQ(e)||UQ(e)||GQ(e)||HQ()}function HQ(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function GQ(e,t){if(e){if(typeof e=="string")return wb(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return wb(e,t)}}function wb(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function UQ(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function WQ(e){if(Array.isArray(e))return e}function qQ(){var e={},t=function(){return null},n=!1,r=function a(o){if(!n){if(Array.isArray(o)){if(!o.length)return;var i=o,s=zQ(i),l=s[0],u=s.slice(1);if(typeof l=="number"){xb(a.bind(null,u),l);return}a(l),xb(a.bind(null,u));return}z0(o)==="object"&&(e=o,t(e)),typeof o=="function"&&o()}};return{stop:function(){n=!0},start:function(o){n=!1,r(o)},subscribe:function(o){return t=o,function(){t=function(){return null}}}}}function Jl(e){"@babel/helpers - typeof";return Jl=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Jl(e)}function bb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Pb(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?bb(Object(n),!0).forEach(function(r){PP(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):bb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function PP(e,t,n){return t=VQ(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function VQ(e){var t=KQ(e,"string");return Jl(t)==="symbol"?t:String(t)}function KQ(e,t){if(Jl(e)!=="object"||e===null)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Jl(r)!=="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var XQ=function(t,n){return[Object.keys(t),Object.keys(n)].reduce(function(r,a){return r.filter(function(o){return a.includes(o)})})},YQ=function(t){return t},QQ=function(t){return t.replace(/([A-Z])/g,function(n){return"-".concat(n.toLowerCase())})},ul=function(t,n){return Object.keys(n).reduce(function(r,a){return Pb(Pb({},r),{},PP({},a,t(a,n[a])))},{})},Sb=function(t,n,r){return t.map(function(a){return"".concat(QQ(a)," ").concat(n,"ms ").concat(r)}).join(",")};function ZQ(e,t){return tZ(e)||eZ(e,t)||SP(e,t)||JQ()}function JQ(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function eZ(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function tZ(e){if(Array.isArray(e))return e}function nZ(e){return oZ(e)||aZ(e)||SP(e)||rZ()}function rZ(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function SP(e,t){if(e){if(typeof e=="string")return H0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return H0(e,t)}}function aZ(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function oZ(e){if(Array.isArray(e))return H0(e)}function H0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}var Kp=1e-4,OP=function(t,n){return[0,3*t,3*n-6*t,3*t-3*n+1]},kP=function(t,n){return t.map(function(r,a){return r*Math.pow(n,a)}).reduce(function(r,a){return r+a})},Ob=function(t,n){return function(r){var a=OP(t,n);return kP(a,r)}},iZ=function(t,n){return function(r){var a=OP(t,n),o=[].concat(nZ(a.map(function(i,s){return i*s}).slice(1)),[0]);return kP(o,r)}},kb=function(){for(var t=arguments.length,n=new Array(t),r=0;r<t;r++)n[r]=arguments[r];var a=n[0],o=n[1],i=n[2],s=n[3];if(n.length===1)switch(n[0]){case"linear":a=0,o=0,i=1,s=1;break;case"ease":a=.25,o=.1,i=.25,s=1;break;case"ease-in":a=.42,o=0,i=1,s=1;break;case"ease-out":a=.42,o=0,i=.58,s=1;break;case"ease-in-out":a=0,o=0,i=.58,s=1;break;default:{var l=n[0].split("(");if(l[0]==="cubic-bezier"&&l[1].split(")")[0].split(",").length===4){var u=l[1].split(")")[0].split(",").map(function(g){return parseFloat(g)}),p=ZQ(u,4);a=p[0],o=p[1],i=p[2],s=p[3]}}}var c=Ob(a,i),f=Ob(o,s),m=iZ(a,i),h=function(v){return v>1?1:v<0?0:v},d=function(v){for(var y=v>1?1:v,x=y,P=0;P<8;++P){var w=c(x)-y,S=m(x);if(Math.abs(w-y)<Kp||S<Kp)return f(x);x=h(x-w/S)}return f(x)};return d.isStepper=!1,d},sZ=function(){var t=arguments.length>0&&arguments[0]!==void 0?arguments[0]:{},n=t.stiff,r=n===void 0?100:n,a=t.damping,o=a===void 0?8:a,i=t.dt,s=i===void 0?17:i,l=function(p,c,f){var m=-(p-c)*r,h=f*o,d=f+(m-h)*s/1e3,g=f*s/1e3+p;return Math.abs(g-c)<Kp&&Math.abs(d)<Kp?[c,0]:[g,d]};return l.isStepper=!0,l.dt=s,l},lZ=function(){for(var t=arguments.length,n=new Array(t),r=0;r<t;r++)n[r]=arguments[r];var a=n[0];if(typeof a=="string")switch(a){case"ease":case"ease-in-out":case"ease-out":case"ease-in":case"linear":return kb(a);case"spring":return sZ();default:if(a.split("(")[0]==="cubic-bezier")return kb(a)}return typeof a=="function"?a:null};function eu(e){"@babel/helpers - typeof";return eu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},eu(e)}function Cb(e){return pZ(e)||cZ(e)||CP(e)||uZ()}function uZ(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function cZ(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function pZ(e){if(Array.isArray(e))return U0(e)}function _b(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function pt(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?_b(Object(n),!0).forEach(function(r){G0(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):_b(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function G0(e,t,n){return t=fZ(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function fZ(e){var t=dZ(e,"string");return eu(t)==="symbol"?t:String(t)}function dZ(e,t){if(eu(e)!=="object"||e===null)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(eu(r)!=="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function mZ(e,t){return yZ(e)||vZ(e,t)||CP(e,t)||hZ()}function hZ(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function CP(e,t){if(e){if(typeof e=="string")return U0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return U0(e,t)}}function U0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function vZ(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function yZ(e){if(Array.isArray(e))return e}var Xp=function(t,n,r){return t+(n-t)*r},W0=function(t){var n=t.from,r=t.to;return n!==r},gZ=function e(t,n,r){var a=ul(function(o,i){if(W0(i)){var s=t(i.from,i.to,i.velocity),l=mZ(s,2),u=l[0],p=l[1];return pt(pt({},i),{},{from:u,velocity:p})}return i},n);return r<1?ul(function(o,i){return W0(i)?pt(pt({},i),{},{velocity:Xp(i.velocity,a[o].velocity,r),from:Xp(i.from,a[o].from,r)}):i},n):e(t,a,r-1)};const xZ=function(e,t,n,r,a){var o=XQ(e,t),i=o.reduce(function(g,v){return pt(pt({},g),{},G0({},v,[e[v],t[v]]))},{}),s=o.reduce(function(g,v){return pt(pt({},g),{},G0({},v,{from:e[v],velocity:0,to:t[v]}))},{}),l=-1,u,p,c=function(){return null},f=function(){return ul(function(v,y){return y.from},s)},m=function(){return!Object.values(s).filter(W0).length},h=function(v){u||(u=v);var y=v-u,x=y/n.dt;s=gZ(n,s,x),a(pt(pt(pt({},e),t),f())),u=v,m()||(l=requestAnimationFrame(c))},d=function(v){p||(p=v);var y=(v-p)/r,x=ul(function(w,S){return Xp.apply(void 0,Cb(S).concat([n(y)]))},i);if(a(pt(pt(pt({},e),t),x)),y<1)l=requestAnimationFrame(c);else{var P=ul(function(w,S){return Xp.apply(void 0,Cb(S).concat([n(1)]))},i);a(pt(pt(pt({},e),t),P))}};return c=n.isStepper?h:d,function(){return requestAnimationFrame(c),function(){cancelAnimationFrame(l)}}};function zi(e){"@babel/helpers - typeof";return zi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},zi(e)}var wZ=["children","begin","duration","attributeName","easing","isActive","steps","from","to","canBegin","onAnimationEnd","shouldReAnimate","onAnimationReStart"];function bZ(e,t){if(e==null)return{};var n=PZ(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function PZ(e,t){if(e==null)return{};var n={},r=Object.keys(e),a,o;for(o=0;o<r.length;o++)a=r[o],!(t.indexOf(a)>=0)&&(n[a]=e[a]);return n}function Dm(e){return CZ(e)||kZ(e)||OZ(e)||SZ()}function SZ(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function OZ(e,t){if(e){if(typeof e=="string")return q0(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return q0(e,t)}}function kZ(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function CZ(e){if(Array.isArray(e))return q0(e)}function q0(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function Ab(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function kn(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Ab(Object(n),!0).forEach(function(r){Ys(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Ab(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function Ys(e,t,n){return t=_P(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function _Z(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function AZ(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,_P(r.key),r)}}function EZ(e,t,n){return t&&AZ(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function _P(e){var t=TZ(e,"string");return zi(t)==="symbol"?t:String(t)}function TZ(e,t){if(zi(e)!=="object"||e===null)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(zi(r)!=="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function jZ(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&V0(e,t)}function V0(e,t){return V0=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},V0(e,t)}function NZ(e){var t=$Z();return function(){var r=Yp(e),a;if(t){var o=Yp(this).constructor;a=Reflect.construct(r,arguments,o)}else a=r.apply(this,arguments);return K0(this,a)}}function K0(e,t){if(t&&(zi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return X0(e)}function X0(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function $Z(){if(typeof Reflect>"u"||!Reflect.construct||Reflect.construct.sham)return!1;if(typeof Proxy=="function")return!0;try{return Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){})),!0}catch{return!1}}function Yp(e){return Yp=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},Yp(e)}var mr=function(e){jZ(n,e);var t=NZ(n);function n(r,a){var o;_Z(this,n),o=t.call(this,r,a);var i=o.props,s=i.isActive,l=i.attributeName,u=i.from,p=i.to,c=i.steps,f=i.children,m=i.duration;if(o.handleStyleChange=o.handleStyleChange.bind(X0(o)),o.changeStyle=o.changeStyle.bind(X0(o)),!s||m<=0)return o.state={style:{}},typeof f=="function"&&(o.state={style:p}),K0(o);if(c&&c.length)o.state={style:c[0].style};else if(u){if(typeof f=="function")return o.state={style:u},K0(o);o.state={style:l?Ys({},l,u):u}}else o.state={style:{}};return o}return EZ(n,[{key:"componentDidMount",value:function(){var a=this.props,o=a.isActive,i=a.canBegin;this.mounted=!0,!(!o||!i)&&this.runAnimation(this.props)}},{key:"componentDidUpdate",value:function(a){var o=this.props,i=o.isActive,s=o.canBegin,l=o.attributeName,u=o.shouldReAnimate,p=o.to,c=o.from,f=this.state.style;if(s){if(!i){var m={style:l?Ys({},l,p):p};this.state&&f&&(l&&f[l]!==p||!l&&f!==p)&&this.setState(m);return}if(!(FQ(a.to,p)&&a.canBegin&&a.isActive)){var h=!a.canBegin||!a.isActive;this.manager&&this.manager.stop(),this.stopJSAnimation&&this.stopJSAnimation();var d=h||u?c:a.to;if(this.state&&f){var g={style:l?Ys({},l,d):d};(l&&f[l]!==d||!l&&f!==d)&&this.setState(g)}this.runAnimation(kn(kn({},this.props),{},{from:d,begin:0}))}}}},{key:"componentWillUnmount",value:function(){this.mounted=!1;var a=this.props.onAnimationEnd;this.unSubscribe&&this.unSubscribe(),this.manager&&(this.manager.stop(),this.manager=null),this.stopJSAnimation&&this.stopJSAnimation(),a&&a()}},{key:"handleStyleChange",value:function(a){this.changeStyle(a)}},{key:"changeStyle",value:function(a){this.mounted&&this.setState({style:a})}},{key:"runJSAnimation",value:function(a){var o=this,i=a.from,s=a.to,l=a.duration,u=a.easing,p=a.begin,c=a.onAnimationEnd,f=a.onAnimationStart,m=xZ(i,s,lZ(u),l,this.changeStyle),h=function(){o.stopJSAnimation=m()};this.manager.start([f,p,h,l,c])}},{key:"runStepAnimation",value:function(a){var o=this,i=a.steps,s=a.begin,l=a.onAnimationStart,u=i[0],p=u.style,c=u.duration,f=c===void 0?0:c,m=function(d,g,v){if(v===0)return d;var y=g.duration,x=g.easing,P=x===void 0?"ease":x,w=g.style,S=g.properties,O=g.onAnimationEnd,C=v>0?i[v-1]:g,_=S||Object.keys(w);if(typeof P=="function"||P==="spring")return[].concat(Dm(d),[o.runJSAnimation.bind(o,{from:C.style,to:w,duration:y,easing:P}),y]);var T=Sb(_,y,P),A=kn(kn(kn({},C.style),w),{},{transition:T});return[].concat(Dm(d),[A,y,O]).filter(YQ)};return this.manager.start([l].concat(Dm(i.reduce(m,[p,Math.max(f,s)])),[a.onAnimationEnd]))}},{key:"runAnimation",value:function(a){this.manager||(this.manager=qQ());var o=a.begin,i=a.duration,s=a.attributeName,l=a.to,u=a.easing,p=a.onAnimationStart,c=a.onAnimationEnd,f=a.steps,m=a.children,h=this.manager;if(this.unSubscribe=h.subscribe(this.handleStyleChange),typeof u=="function"||typeof m=="function"||u==="spring"){this.runJSAnimation(a);return}if(f.length>1){this.runStepAnimation(a);return}var d=s?Ys({},s,l):l,g=Sb(Object.keys(d),i,u);h.start([p,o,kn(kn({},d),{},{transition:g}),i,c])}},{key:"render",value:function(){var a=this.props,o=a.children;a.begin;var i=a.duration;a.attributeName,a.easing;var s=a.isActive;a.steps,a.from,a.to,a.canBegin,a.onAnimationEnd,a.shouldReAnimate,a.onAnimationReStart;var l=bZ(a,wZ),u=k.Children.count(o),p=this.state.style;if(typeof o=="function")return o(p);if(!s||u===0||i<=0)return o;var c=function(m){var h=m.props,d=h.style,g=d===void 0?{}:d,v=h.className,y=k.cloneElement(m,kn(kn({},l),{},{style:kn(kn({},g),p),className:v}));return y};return u===1?c(k.Children.only(o)):E.createElement("div",null,k.Children.map(o,function(f){return c(f)}))}}]),n}(k.PureComponent);mr.displayName="Animate";mr.defaultProps={begin:0,duration:1e3,from:"",to:"",attributeName:"",easing:"ease",isActive:!0,canBegin:!0,steps:[],onAnimationEnd:function(){},onAnimationStart:function(){}};mr.propTypes={from:ye.oneOfType([ye.object,ye.string]),to:ye.oneOfType([ye.object,ye.string]),attributeName:ye.string,duration:ye.number,begin:ye.number,easing:ye.oneOfType([ye.string,ye.func]),steps:ye.arrayOf(ye.shape({duration:ye.number.isRequired,style:ye.object.isRequired,easing:ye.oneOfType([ye.oneOf(["ease","ease-in","ease-out","ease-in-out","linear"]),ye.func]),properties:ye.arrayOf("string"),onAnimationEnd:ye.func})),children:ye.oneOfType([ye.node,ye.func]),isActive:ye.bool,canBegin:ye.bool,onAnimationEnd:ye.func,shouldReAnimate:ye.bool,onAnimationStart:ye.func,onAnimationReStart:ye.func};function tu(e){"@babel/helpers - typeof";return tu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},tu(e)}function Qp(){return Qp=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Qp.apply(this,arguments)}function MZ(e,t){return LZ(e)||DZ(e,t)||IZ(e,t)||RZ()}function RZ(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function IZ(e,t){if(e){if(typeof e=="string")return Eb(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return Eb(e,t)}}function Eb(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function DZ(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function LZ(e){if(Array.isArray(e))return e}function Tb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function jb(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Tb(Object(n),!0).forEach(function(r){FZ(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Tb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function FZ(e,t,n){return t=BZ(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function BZ(e){var t=zZ(e,"string");return tu(t)=="symbol"?t:t+""}function zZ(e,t){if(tu(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(tu(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var Nb=function(t,n,r,a,o){var i=Math.min(Math.abs(r)/2,Math.abs(a)/2),s=a>=0?1:-1,l=r>=0?1:-1,u=a>=0&&r>=0||a<0&&r<0?1:0,p;if(i>0&&o instanceof Array){for(var c=[0,0,0,0],f=0,m=4;f<m;f++)c[f]=o[f]>i?i:o[f];p="M".concat(t,",").concat(n+s*c[0]),c[0]>0&&(p+="A ".concat(c[0],",").concat(c[0],",0,0,").concat(u,",").concat(t+l*c[0],",").concat(n)),p+="L ".concat(t+r-l*c[1],",").concat(n),c[1]>0&&(p+="A ".concat(c[1],",").concat(c[1],",0,0,").concat(u,`,
        `).concat(t+r,",").concat(n+s*c[1])),p+="L ".concat(t+r,",").concat(n+a-s*c[2]),c[2]>0&&(p+="A ".concat(c[2],",").concat(c[2],",0,0,").concat(u,`,
        `).concat(t+r-l*c[2],",").concat(n+a)),p+="L ".concat(t+l*c[3],",").concat(n+a),c[3]>0&&(p+="A ".concat(c[3],",").concat(c[3],",0,0,").concat(u,`,
        `).concat(t,",").concat(n+a-s*c[3])),p+="Z"}else if(i>0&&o===+o&&o>0){var h=Math.min(i,o);p="M ".concat(t,",").concat(n+s*h,`
            A `).concat(h,",").concat(h,",0,0,").concat(u,",").concat(t+l*h,",").concat(n,`
            L `).concat(t+r-l*h,",").concat(n,`
            A `).concat(h,",").concat(h,",0,0,").concat(u,",").concat(t+r,",").concat(n+s*h,`
            L `).concat(t+r,",").concat(n+a-s*h,`
            A `).concat(h,",").concat(h,",0,0,").concat(u,",").concat(t+r-l*h,",").concat(n+a,`
            L `).concat(t+l*h,",").concat(n+a,`
            A `).concat(h,",").concat(h,",0,0,").concat(u,",").concat(t,",").concat(n+a-s*h," Z")}else p="M ".concat(t,",").concat(n," h ").concat(r," v ").concat(a," h ").concat(-r," Z");return p},HZ=function(t,n){if(!t||!n)return!1;var r=t.x,a=t.y,o=n.x,i=n.y,s=n.width,l=n.height;if(Math.abs(s)>0&&Math.abs(l)>0){var u=Math.min(o,o+s),p=Math.max(o,o+s),c=Math.min(i,i+l),f=Math.max(i,i+l);return r>=u&&r<=p&&a>=c&&a<=f}return!1},GZ={x:0,y:0,width:0,height:0,radius:0,isAnimationActive:!1,isUpdateAnimationActive:!1,animationBegin:0,animationDuration:1500,animationEasing:"ease"},Ag=function(t){var n=jb(jb({},GZ),t),r=k.useRef(),a=k.useState(-1),o=MZ(a,2),i=o[0],s=o[1];k.useEffect(function(){if(r.current&&r.current.getTotalLength)try{var P=r.current.getTotalLength();P&&s(P)}catch{}},[]);var l=n.x,u=n.y,p=n.width,c=n.height,f=n.radius,m=n.className,h=n.animationEasing,d=n.animationDuration,g=n.animationBegin,v=n.isAnimationActive,y=n.isUpdateAnimationActive;if(l!==+l||u!==+u||p!==+p||c!==+c||p===0||c===0)return null;var x=ue("recharts-rectangle",m);return y?E.createElement(mr,{canBegin:i>0,from:{width:p,height:c,x:l,y:u},to:{width:p,height:c,x:l,y:u},duration:d,animationEasing:h,isActive:y},function(P){var w=P.width,S=P.height,O=P.x,C=P.y;return E.createElement(mr,{canBegin:i>0,from:"0px ".concat(i===-1?1:i,"px"),to:"".concat(i,"px 0px"),attributeName:"strokeDasharray",begin:g,duration:d,isActive:v,easing:h},E.createElement("path",Qp({},ie(n,!0),{className:x,d:Nb(O,C,w,S,f),ref:r})))}):E.createElement("path",Qp({},ie(n,!0),{className:x,d:Nb(l,u,p,c,f)}))};function Y0(){return Y0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Y0.apply(this,arguments)}var gd=function(t){var n=t.cx,r=t.cy,a=t.r,o=t.className,i=ue("recharts-dot",o);return n===+n&&r===+r&&a===+a?k.createElement("circle",Y0({},ie(t,!1),hp(t),{className:i,cx:n,cy:r,r:a})):null};function nu(e){"@babel/helpers - typeof";return nu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},nu(e)}var UZ=["x","y","top","left","width","height","className"];function Q0(){return Q0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Q0.apply(this,arguments)}function $b(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function WZ(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?$b(Object(n),!0).forEach(function(r){qZ(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):$b(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function qZ(e,t,n){return t=VZ(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function VZ(e){var t=KZ(e,"string");return nu(t)=="symbol"?t:t+""}function KZ(e,t){if(nu(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(nu(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function XZ(e,t){if(e==null)return{};var n=YZ(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function YZ(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}var QZ=function(t,n,r,a,o,i){return"M".concat(t,",").concat(o,"v").concat(a,"M").concat(i,",").concat(n,"h").concat(r)},ZZ=function(t){var n=t.x,r=n===void 0?0:n,a=t.y,o=a===void 0?0:a,i=t.top,s=i===void 0?0:i,l=t.left,u=l===void 0?0:l,p=t.width,c=p===void 0?0:p,f=t.height,m=f===void 0?0:f,h=t.className,d=XZ(t,UZ),g=WZ({x:r,y:o,top:s,left:u,width:c,height:m},d);return!V(r)||!V(o)||!V(c)||!V(m)||!V(s)||!V(u)?null:E.createElement("path",Q0({},ie(g,!0),{className:ue("recharts-cross",h),d:QZ(r,o,c,m,s,u)}))},JZ=z8,eJ=JZ(Object.getPrototypeOf,Object),tJ=eJ,nJ=Gr,rJ=tJ,aJ=Ur,oJ="[object Object]",iJ=Function.prototype,sJ=Object.prototype,AP=iJ.toString,lJ=sJ.hasOwnProperty,uJ=AP.call(Object);function cJ(e){if(!aJ(e)||nJ(e)!=oJ)return!1;var t=rJ(e);if(t===null)return!0;var n=lJ.call(t,"constructor")&&t.constructor;return typeof n=="function"&&n instanceof n&&AP.call(n)==uJ}var pJ=cJ;const fJ=_e(pJ);var dJ=Gr,mJ=Ur,hJ="[object Boolean]";function vJ(e){return e===!0||e===!1||mJ(e)&&dJ(e)==hJ}var yJ=vJ;const gJ=_e(yJ);function ru(e){"@babel/helpers - typeof";return ru=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},ru(e)}function Zp(){return Zp=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Zp.apply(this,arguments)}function xJ(e,t){return SJ(e)||PJ(e,t)||bJ(e,t)||wJ()}function wJ(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function bJ(e,t){if(e){if(typeof e=="string")return Mb(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return Mb(e,t)}}function Mb(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function PJ(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function SJ(e){if(Array.isArray(e))return e}function Rb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Ib(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Rb(Object(n),!0).forEach(function(r){OJ(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Rb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function OJ(e,t,n){return t=kJ(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function kJ(e){var t=CJ(e,"string");return ru(t)=="symbol"?t:t+""}function CJ(e,t){if(ru(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(ru(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var Db=function(t,n,r,a,o){var i=r-a,s;return s="M ".concat(t,",").concat(n),s+="L ".concat(t+r,",").concat(n),s+="L ".concat(t+r-i/2,",").concat(n+o),s+="L ".concat(t+r-i/2-a,",").concat(n+o),s+="L ".concat(t,",").concat(n," Z"),s},_J={x:0,y:0,upperWidth:0,lowerWidth:0,height:0,isUpdateAnimationActive:!1,animationBegin:0,animationDuration:1500,animationEasing:"ease"},AJ=function(t){var n=Ib(Ib({},_J),t),r=k.useRef(),a=k.useState(-1),o=xJ(a,2),i=o[0],s=o[1];k.useEffect(function(){if(r.current&&r.current.getTotalLength)try{var x=r.current.getTotalLength();x&&s(x)}catch{}},[]);var l=n.x,u=n.y,p=n.upperWidth,c=n.lowerWidth,f=n.height,m=n.className,h=n.animationEasing,d=n.animationDuration,g=n.animationBegin,v=n.isUpdateAnimationActive;if(l!==+l||u!==+u||p!==+p||c!==+c||f!==+f||p===0&&c===0||f===0)return null;var y=ue("recharts-trapezoid",m);return v?E.createElement(mr,{canBegin:i>0,from:{upperWidth:0,lowerWidth:0,height:f,x:l,y:u},to:{upperWidth:p,lowerWidth:c,height:f,x:l,y:u},duration:d,animationEasing:h,isActive:v},function(x){var P=x.upperWidth,w=x.lowerWidth,S=x.height,O=x.x,C=x.y;return E.createElement(mr,{canBegin:i>0,from:"0px ".concat(i===-1?1:i,"px"),to:"".concat(i,"px 0px"),attributeName:"strokeDasharray",begin:g,duration:d,easing:h},E.createElement("path",Zp({},ie(n,!0),{className:y,d:Db(O,C,P,w,S),ref:r})))}):E.createElement("g",null,E.createElement("path",Zp({},ie(n,!0),{className:y,d:Db(l,u,p,c,f)})))},EJ=["option","shapeType","propTransformer","activeClassName","isActive"];function au(e){"@babel/helpers - typeof";return au=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},au(e)}function TJ(e,t){if(e==null)return{};var n=jJ(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function jJ(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function Lb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Jp(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Lb(Object(n),!0).forEach(function(r){NJ(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Lb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function NJ(e,t,n){return t=$J(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function $J(e){var t=MJ(e,"string");return au(t)=="symbol"?t:t+""}function MJ(e,t){if(au(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(au(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function RJ(e,t){return Jp(Jp({},t),e)}function IJ(e,t){return e==="symbols"}function Fb(e){var t=e.shapeType,n=e.elementProps;switch(t){case"rectangle":return E.createElement(Ag,n);case"trapezoid":return E.createElement(AJ,n);case"sector":return E.createElement(yP,n);case"symbols":if(IJ(t))return E.createElement(Vy,n);break;default:return null}}function DJ(e){return k.isValidElement(e)?e.props:e}function LJ(e){var t=e.option,n=e.shapeType,r=e.propTransformer,a=r===void 0?RJ:r,o=e.activeClassName,i=o===void 0?"recharts-active-shape":o,s=e.isActive,l=TJ(e,EJ),u;if(k.isValidElement(t))u=k.cloneElement(t,Jp(Jp({},l),DJ(t)));else if(ae(t))u=t(l);else if(fJ(t)&&!gJ(t)){var p=a(t,l);u=E.createElement(Fb,{shapeType:n,elementProps:p})}else{var c=l;u=E.createElement(Fb,{shapeType:n,elementProps:c})}return s?E.createElement(Ne,{className:i},u):u}function xd(e,t){return t!=null&&"trapezoids"in e.props}function wd(e,t){return t!=null&&"sectors"in e.props}function ou(e,t){return t!=null&&"points"in e.props}function FJ(e,t){var n,r,a=e.x===(t==null||(n=t.labelViewBox)===null||n===void 0?void 0:n.x)||e.x===t.x,o=e.y===(t==null||(r=t.labelViewBox)===null||r===void 0?void 0:r.y)||e.y===t.y;return a&&o}function BJ(e,t){var n=e.endAngle===t.endAngle,r=e.startAngle===t.startAngle;return n&&r}function zJ(e,t){var n=e.x===t.x,r=e.y===t.y,a=e.z===t.z;return n&&r&&a}function HJ(e,t){var n;return xd(e,t)?n=FJ:wd(e,t)?n=BJ:ou(e,t)&&(n=zJ),n}function GJ(e,t){var n;return xd(e,t)?n="trapezoids":wd(e,t)?n="sectors":ou(e,t)&&(n="points"),n}function UJ(e,t){if(xd(e,t)){var n;return(n=t.tooltipPayload)===null||n===void 0||(n=n[0])===null||n===void 0||(n=n.payload)===null||n===void 0?void 0:n.payload}if(wd(e,t)){var r;return(r=t.tooltipPayload)===null||r===void 0||(r=r[0])===null||r===void 0||(r=r.payload)===null||r===void 0?void 0:r.payload}return ou(e,t)?t.payload:{}}function WJ(e){var t=e.activeTooltipItem,n=e.graphicalItem,r=e.itemData,a=GJ(n,t),o=UJ(n,t),i=r.filter(function(l,u){var p=Fi(o,l),c=n.props[a].filter(function(h){var d=HJ(n,t);return d(h,t)}),f=n.props[a].indexOf(c[c.length-1]),m=u===f;return p&&m}),s=r.indexOf(i[i.length-1]);return s}var qJ=Math.ceil,VJ=Math.max;function KJ(e,t,n,r){for(var a=-1,o=VJ(qJ((t-e)/(n||1)),0),i=Array(o);o--;)i[r?o:++a]=e,e+=n;return i}var XJ=KJ,YJ=i7,Bb=1/0,QJ=17976931348623157e292;function ZJ(e){if(!e)return e===0?e:0;if(e=YJ(e),e===Bb||e===-Bb){var t=e<0?-1:1;return t*QJ}return e===e?e:0}var EP=ZJ,JJ=XJ,eee=ld,Lm=EP;function tee(e){return function(t,n,r){return r&&typeof r!="number"&&eee(t,n,r)&&(n=r=void 0),t=Lm(t),n===void 0?(n=t,t=0):n=Lm(n),r=r===void 0?t<n?1:-1:Lm(r),JJ(t,n,r,e)}}var nee=tee,ree=nee,aee=ree(),oee=aee;const ef=_e(oee);function iu(e){"@babel/helpers - typeof";return iu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},iu(e)}function zb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Hb(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?zb(Object(n),!0).forEach(function(r){TP(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):zb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function TP(e,t,n){return t=iee(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function iee(e){var t=see(e,"string");return iu(t)=="symbol"?t:t+""}function see(e,t){if(iu(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(iu(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var lee=["Webkit","Moz","O","ms"],uee=function(t,n){var r=t.replace(/(\w)/,function(o){return o.toUpperCase()}),a=lee.reduce(function(o,i){return Hb(Hb({},o),{},TP({},i+r,n))},{});return a[t]=n,a};function Hi(e){"@babel/helpers - typeof";return Hi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Hi(e)}function tf(){return tf=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},tf.apply(this,arguments)}function Gb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Fm(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Gb(Object(n),!0).forEach(function(r){Kt(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Gb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function cee(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Ub(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,NP(r.key),r)}}function pee(e,t,n){return t&&Ub(e.prototype,t),n&&Ub(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function fee(e,t,n){return t=nf(t),dee(e,jP()?Reflect.construct(t,n||[],nf(e).constructor):t.apply(e,n))}function dee(e,t){if(t&&(Hi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return mee(e)}function mee(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function jP(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(jP=function(){return!!e})()}function nf(e){return nf=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},nf(e)}function hee(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&Z0(e,t)}function Z0(e,t){return Z0=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},Z0(e,t)}function Kt(e,t,n){return t=NP(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function NP(e){var t=vee(e,"string");return Hi(t)=="symbol"?t:t+""}function vee(e,t){if(Hi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Hi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var yee=function(t){var n=t.data,r=t.startIndex,a=t.endIndex,o=t.x,i=t.width,s=t.travellerWidth;if(!n||!n.length)return{};var l=n.length,u=sl().domain(ef(0,l)).range([o,o+i-s]),p=u.domain().map(function(c){return u(c)});return{isTextActive:!1,isSlideMoving:!1,isTravellerMoving:!1,isTravellerFocused:!1,startX:u(r),endX:u(a),scale:u,scaleValues:p}},Wb=function(t){return t.changedTouches&&!!t.changedTouches.length},Gi=function(e){function t(n){var r;return cee(this,t),r=fee(this,t,[n]),Kt(r,"handleDrag",function(a){r.leaveTimer&&(clearTimeout(r.leaveTimer),r.leaveTimer=null),r.state.isTravellerMoving?r.handleTravellerMove(a):r.state.isSlideMoving&&r.handleSlideDrag(a)}),Kt(r,"handleTouchMove",function(a){a.changedTouches!=null&&a.changedTouches.length>0&&r.handleDrag(a.changedTouches[0])}),Kt(r,"handleDragEnd",function(){r.setState({isTravellerMoving:!1,isSlideMoving:!1},function(){var a=r.props,o=a.endIndex,i=a.onDragEnd,s=a.startIndex;i==null||i({endIndex:o,startIndex:s})}),r.detachDragEndListener()}),Kt(r,"handleLeaveWrapper",function(){(r.state.isTravellerMoving||r.state.isSlideMoving)&&(r.leaveTimer=window.setTimeout(r.handleDragEnd,r.props.leaveTimeOut))}),Kt(r,"handleEnterSlideOrTraveller",function(){r.setState({isTextActive:!0})}),Kt(r,"handleLeaveSlideOrTraveller",function(){r.setState({isTextActive:!1})}),Kt(r,"handleSlideDragStart",function(a){var o=Wb(a)?a.changedTouches[0]:a;r.setState({isTravellerMoving:!1,isSlideMoving:!0,slideMoveStartX:o.pageX}),r.attachDragEndListener()}),r.travellerDragStartHandlers={startX:r.handleTravellerDragStart.bind(r,"startX"),endX:r.handleTravellerDragStart.bind(r,"endX")},r.state={},r}return hee(t,e),pee(t,[{key:"componentWillUnmount",value:function(){this.leaveTimer&&(clearTimeout(this.leaveTimer),this.leaveTimer=null),this.detachDragEndListener()}},{key:"getIndex",value:function(r){var a=r.startX,o=r.endX,i=this.state.scaleValues,s=this.props,l=s.gap,u=s.data,p=u.length-1,c=Math.min(a,o),f=Math.max(a,o),m=t.getIndexInRange(i,c),h=t.getIndexInRange(i,f);return{startIndex:m-m%l,endIndex:h===p?p:h-h%l}}},{key:"getTextOfTick",value:function(r){var a=this.props,o=a.data,i=a.tickFormatter,s=a.dataKey,l=It(o[r],s,r);return ae(i)?i(l,r):l}},{key:"attachDragEndListener",value:function(){window.addEventListener("mouseup",this.handleDragEnd,!0),window.addEventListener("touchend",this.handleDragEnd,!0),window.addEventListener("mousemove",this.handleDrag,!0)}},{key:"detachDragEndListener",value:function(){window.removeEventListener("mouseup",this.handleDragEnd,!0),window.removeEventListener("touchend",this.handleDragEnd,!0),window.removeEventListener("mousemove",this.handleDrag,!0)}},{key:"handleSlideDrag",value:function(r){var a=this.state,o=a.slideMoveStartX,i=a.startX,s=a.endX,l=this.props,u=l.x,p=l.width,c=l.travellerWidth,f=l.startIndex,m=l.endIndex,h=l.onChange,d=r.pageX-o;d>0?d=Math.min(d,u+p-c-s,u+p-c-i):d<0&&(d=Math.max(d,u-i,u-s));var g=this.getIndex({startX:i+d,endX:s+d});(g.startIndex!==f||g.endIndex!==m)&&h&&h(g),this.setState({startX:i+d,endX:s+d,slideMoveStartX:r.pageX})}},{key:"handleTravellerDragStart",value:function(r,a){var o=Wb(a)?a.changedTouches[0]:a;this.setState({isSlideMoving:!1,isTravellerMoving:!0,movingTravellerId:r,brushMoveStartX:o.pageX}),this.attachDragEndListener()}},{key:"handleTravellerMove",value:function(r){var a=this.state,o=a.brushMoveStartX,i=a.movingTravellerId,s=a.endX,l=a.startX,u=this.state[i],p=this.props,c=p.x,f=p.width,m=p.travellerWidth,h=p.onChange,d=p.gap,g=p.data,v={startX:this.state.startX,endX:this.state.endX},y=r.pageX-o;y>0?y=Math.min(y,c+f-m-u):y<0&&(y=Math.max(y,c-u)),v[i]=u+y;var x=this.getIndex(v),P=x.startIndex,w=x.endIndex,S=function(){var C=g.length-1;return i==="startX"&&(s>l?P%d===0:w%d===0)||s<l&&w===C||i==="endX"&&(s>l?w%d===0:P%d===0)||s>l&&w===C};this.setState(Kt(Kt({},i,u+y),"brushMoveStartX",r.pageX),function(){h&&S()&&h(x)})}},{key:"handleTravellerMoveKeyboard",value:function(r,a){var o=this,i=this.state,s=i.scaleValues,l=i.startX,u=i.endX,p=this.state[a],c=s.indexOf(p);if(c!==-1){var f=c+r;if(!(f===-1||f>=s.length)){var m=s[f];a==="startX"&&m>=u||a==="endX"&&m<=l||this.setState(Kt({},a,m),function(){o.props.onChange(o.getIndex({startX:o.state.startX,endX:o.state.endX}))})}}}},{key:"renderBackground",value:function(){var r=this.props,a=r.x,o=r.y,i=r.width,s=r.height,l=r.fill,u=r.stroke;return E.createElement("rect",{stroke:u,fill:l,x:a,y:o,width:i,height:s})}},{key:"renderPanorama",value:function(){var r=this.props,a=r.x,o=r.y,i=r.width,s=r.height,l=r.data,u=r.children,p=r.padding,c=k.Children.only(u);return c?E.cloneElement(c,{x:a,y:o,width:i,height:s,margin:p,compact:!0,data:l}):null}},{key:"renderTravellerLayer",value:function(r,a){var o,i,s=this,l=this.props,u=l.y,p=l.travellerWidth,c=l.height,f=l.traveller,m=l.ariaLabel,h=l.data,d=l.startIndex,g=l.endIndex,v=Math.max(r,this.props.x),y=Fm(Fm({},ie(this.props,!1)),{},{x:v,y:u,width:p,height:c}),x=m||"Min value: ".concat((o=h[d])===null||o===void 0?void 0:o.name,", Max value: ").concat((i=h[g])===null||i===void 0?void 0:i.name);return E.createElement(Ne,{tabIndex:0,role:"slider","aria-label":x,"aria-valuenow":r,className:"recharts-brush-traveller",onMouseEnter:this.handleEnterSlideOrTraveller,onMouseLeave:this.handleLeaveSlideOrTraveller,onMouseDown:this.travellerDragStartHandlers[a],onTouchStart:this.travellerDragStartHandlers[a],onKeyDown:function(w){["ArrowLeft","ArrowRight"].includes(w.key)&&(w.preventDefault(),w.stopPropagation(),s.handleTravellerMoveKeyboard(w.key==="ArrowRight"?1:-1,a))},onFocus:function(){s.setState({isTravellerFocused:!0})},onBlur:function(){s.setState({isTravellerFocused:!1})},style:{cursor:"col-resize"}},t.renderTraveller(f,y))}},{key:"renderSlide",value:function(r,a){var o=this.props,i=o.y,s=o.height,l=o.stroke,u=o.travellerWidth,p=Math.min(r,a)+u,c=Math.max(Math.abs(a-r)-u,0);return E.createElement("rect",{className:"recharts-brush-slide",onMouseEnter:this.handleEnterSlideOrTraveller,onMouseLeave:this.handleLeaveSlideOrTraveller,onMouseDown:this.handleSlideDragStart,onTouchStart:this.handleSlideDragStart,style:{cursor:"move"},stroke:"none",fill:l,fillOpacity:.2,x:p,y:i,width:c,height:s})}},{key:"renderText",value:function(){var r=this.props,a=r.startIndex,o=r.endIndex,i=r.y,s=r.height,l=r.travellerWidth,u=r.stroke,p=this.state,c=p.startX,f=p.endX,m=5,h={pointerEvents:"none",fill:u};return E.createElement(Ne,{className:"recharts-brush-texts"},E.createElement(Tp,tf({textAnchor:"end",verticalAnchor:"middle",x:Math.min(c,f)-m,y:i+s/2},h),this.getTextOfTick(a)),E.createElement(Tp,tf({textAnchor:"start",verticalAnchor:"middle",x:Math.max(c,f)+l+m,y:i+s/2},h),this.getTextOfTick(o)))}},{key:"render",value:function(){var r=this.props,a=r.data,o=r.className,i=r.children,s=r.x,l=r.y,u=r.width,p=r.height,c=r.alwaysShowText,f=this.state,m=f.startX,h=f.endX,d=f.isTextActive,g=f.isSlideMoving,v=f.isTravellerMoving,y=f.isTravellerFocused;if(!a||!a.length||!V(s)||!V(l)||!V(u)||!V(p)||u<=0||p<=0)return null;var x=ue("recharts-brush",o),P=E.Children.count(i)===1,w=uee("userSelect","none");return E.createElement(Ne,{className:x,onMouseLeave:this.handleLeaveWrapper,onTouchMove:this.handleTouchMove,style:w},this.renderBackground(),P&&this.renderPanorama(),this.renderSlide(m,h),this.renderTravellerLayer(m,"startX"),this.renderTravellerLayer(h,"endX"),(d||g||v||y||c)&&this.renderText())}}],[{key:"renderDefaultTraveller",value:function(r){var a=r.x,o=r.y,i=r.width,s=r.height,l=r.stroke,u=Math.floor(o+s/2)-1;return E.createElement(E.Fragment,null,E.createElement("rect",{x:a,y:o,width:i,height:s,fill:l,stroke:"none"}),E.createElement("line",{x1:a+1,y1:u,x2:a+i-1,y2:u,fill:"none",stroke:"#fff"}),E.createElement("line",{x1:a+1,y1:u+2,x2:a+i-1,y2:u+2,fill:"none",stroke:"#fff"}))}},{key:"renderTraveller",value:function(r,a){var o;return E.isValidElement(r)?o=E.cloneElement(r,a):ae(r)?o=r(a):o=t.renderDefaultTraveller(a),o}},{key:"getDerivedStateFromProps",value:function(r,a){var o=r.data,i=r.width,s=r.x,l=r.travellerWidth,u=r.updateId,p=r.startIndex,c=r.endIndex;if(o!==a.prevData||u!==a.prevUpdateId)return Fm({prevData:o,prevTravellerWidth:l,prevUpdateId:u,prevX:s,prevWidth:i},o&&o.length?yee({data:o,width:i,x:s,travellerWidth:l,startIndex:p,endIndex:c}):{scale:null,scaleValues:null});if(a.scale&&(i!==a.prevWidth||s!==a.prevX||l!==a.prevTravellerWidth)){a.scale.range([s,s+i-l]);var f=a.scale.domain().map(function(m){return a.scale(m)});return{prevData:o,prevTravellerWidth:l,prevUpdateId:u,prevX:s,prevWidth:i,startX:a.scale(r.startIndex),endX:a.scale(r.endIndex),scaleValues:f}}return null}},{key:"getIndexInRange",value:function(r,a){for(var o=r.length,i=0,s=o-1;s-i>1;){var l=Math.floor((i+s)/2);r[l]>a?s=l:i=l}return a>=r[s]?s:i}}])}(k.PureComponent);Kt(Gi,"displayName","Brush");Kt(Gi,"defaultProps",{height:40,travellerWidth:5,gap:1,fill:"#fff",stroke:"#666",padding:{top:1,right:1,bottom:1,left:1},leaveTimeOut:1e3,alwaysShowText:!1});var gee=eg;function xee(e,t){var n;return gee(e,function(r,a,o){return n=t(r,a,o),!n}),!!n}var wee=xee,bee=$8,Pee=$a,See=wee,Oee=qt,kee=ld;function Cee(e,t,n){var r=Oee(e)?bee:See;return n&&kee(e,t,n)&&(t=void 0),r(e,Pee(t))}var _ee=Cee;const Aee=_e(_ee);var pr=function(t,n){var r=t.alwaysShow,a=t.ifOverflow;return r&&(a="extendDomain"),a===n},qb=t7;function Eee(e,t,n){t=="__proto__"&&qb?qb(e,t,{configurable:!0,enumerable:!0,value:n,writable:!0}):e[t]=n}var Tee=Eee,jee=Tee,Nee=J8,$ee=$a;function Mee(e,t){var n={};return t=$ee(t),Nee(e,function(r,a,o){jee(n,a,t(r,a,o))}),n}var Ree=Mee;const Iee=_e(Ree);function Dee(e,t){for(var n=-1,r=e==null?0:e.length;++n<r;)if(!t(e[n],n,e))return!1;return!0}var Lee=Dee,Fee=eg;function Bee(e,t){var n=!0;return Fee(e,function(r,a,o){return n=!!t(r,a,o),n}),n}var zee=Bee,Hee=Lee,Gee=zee,Uee=$a,Wee=qt,qee=ld;function Vee(e,t,n){var r=Wee(e)?Hee:Gee;return n&&qee(e,t,n)&&(t=void 0),r(e,Uee(t))}var Kee=Vee;const $P=_e(Kee);var Xee=["x","y"];function su(e){"@babel/helpers - typeof";return su=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},su(e)}function J0(){return J0=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},J0.apply(this,arguments)}function Vb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Us(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Vb(Object(n),!0).forEach(function(r){Yee(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Vb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function Yee(e,t,n){return t=Qee(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function Qee(e){var t=Zee(e,"string");return su(t)=="symbol"?t:t+""}function Zee(e,t){if(su(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(su(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function Jee(e,t){if(e==null)return{};var n=ete(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function ete(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function tte(e,t){var n=e.x,r=e.y,a=Jee(e,Xee),o="".concat(n),i=parseInt(o,10),s="".concat(r),l=parseInt(s,10),u="".concat(t.height||a.height),p=parseInt(u,10),c="".concat(t.width||a.width),f=parseInt(c,10);return Us(Us(Us(Us(Us({},t),a),i?{x:i}:{}),l?{y:l}:{}),{},{height:p,width:f,name:t.name,radius:t.radius})}function Kb(e){return E.createElement(LJ,J0({shapeType:"rectangle",propTransformer:tte,activeClassName:"recharts-active-bar"},e))}var nte=function(t){var n=arguments.length>1&&arguments[1]!==void 0?arguments[1]:0;return function(r,a){if(typeof t=="number")return t;var o=V(r)||xM(r);return o?t(r,a):(o||Po(),n)}},rte=["value","background"],MP;function Ui(e){"@babel/helpers - typeof";return Ui=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Ui(e)}function ate(e,t){if(e==null)return{};var n=ote(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function ote(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function rf(){return rf=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},rf.apply(this,arguments)}function Xb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Ze(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Xb(Object(n),!0).forEach(function(r){fa(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Xb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function ite(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Yb(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,IP(r.key),r)}}function ste(e,t,n){return t&&Yb(e.prototype,t),n&&Yb(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function lte(e,t,n){return t=af(t),ute(e,RP()?Reflect.construct(t,n||[],af(e).constructor):t.apply(e,n))}function ute(e,t){if(t&&(Ui(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return cte(e)}function cte(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function RP(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(RP=function(){return!!e})()}function af(e){return af=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},af(e)}function pte(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&ev(e,t)}function ev(e,t){return ev=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},ev(e,t)}function fa(e,t,n){return t=IP(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function IP(e){var t=fte(e,"string");return Ui(t)=="symbol"?t:t+""}function fte(e,t){if(Ui(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Ui(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var Da=function(e){function t(){var n;ite(this,t);for(var r=arguments.length,a=new Array(r),o=0;o<r;o++)a[o]=arguments[o];return n=lte(this,t,[].concat(a)),fa(n,"state",{isAnimationFinished:!1}),fa(n,"id",ys("recharts-bar-")),fa(n,"handleAnimationEnd",function(){var i=n.props.onAnimationEnd;n.setState({isAnimationFinished:!0}),i&&i()}),fa(n,"handleAnimationStart",function(){var i=n.props.onAnimationStart;n.setState({isAnimationFinished:!1}),i&&i()}),n}return pte(t,e),ste(t,[{key:"renderRectanglesStatically",value:function(r){var a=this,o=this.props,i=o.shape,s=o.dataKey,l=o.activeIndex,u=o.activeBar,p=ie(this.props,!1);return r&&r.map(function(c,f){var m=f===l,h=m?u:i,d=Ze(Ze(Ze({},p),c),{},{isActive:m,option:h,index:f,dataKey:s,onAnimationStart:a.handleAnimationStart,onAnimationEnd:a.handleAnimationEnd});return E.createElement(Ne,rf({className:"recharts-bar-rectangle"},vp(a.props,c,f),{key:"rectangle-".concat(c==null?void 0:c.x,"-").concat(c==null?void 0:c.y,"-").concat(c==null?void 0:c.value,"-").concat(f)}),E.createElement(Kb,d))})}},{key:"renderRectanglesWithAnimation",value:function(){var r=this,a=this.props,o=a.data,i=a.layout,s=a.isAnimationActive,l=a.animationBegin,u=a.animationDuration,p=a.animationEasing,c=a.animationId,f=this.state.prevData;return E.createElement(mr,{begin:l,duration:u,isActive:s,easing:p,from:{t:0},to:{t:1},key:"bar-".concat(c),onAnimationEnd:this.handleAnimationEnd,onAnimationStart:this.handleAnimationStart},function(m){var h=m.t,d=o.map(function(g,v){var y=f&&f[v];if(y){var x=dt(y.x,g.x),P=dt(y.y,g.y),w=dt(y.width,g.width),S=dt(y.height,g.height);return Ze(Ze({},g),{},{x:x(h),y:P(h),width:w(h),height:S(h)})}if(i==="horizontal"){var O=dt(0,g.height),C=O(h);return Ze(Ze({},g),{},{y:g.y+g.height-C,height:C})}var _=dt(0,g.width),T=_(h);return Ze(Ze({},g),{},{width:T})});return E.createElement(Ne,null,r.renderRectanglesStatically(d))})}},{key:"renderRectangles",value:function(){var r=this.props,a=r.data,o=r.isAnimationActive,i=this.state.prevData;return o&&a&&a.length&&(!i||!Fi(i,a))?this.renderRectanglesWithAnimation():this.renderRectanglesStatically(a)}},{key:"renderBackground",value:function(){var r=this,a=this.props,o=a.data,i=a.dataKey,s=a.activeIndex,l=ie(this.props.background,!1);return o.map(function(u,p){u.value;var c=u.background,f=ate(u,rte);if(!c)return null;var m=Ze(Ze(Ze(Ze(Ze({},f),{},{fill:"#eee"},c),l),vp(r.props,u,p)),{},{onAnimationStart:r.handleAnimationStart,onAnimationEnd:r.handleAnimationEnd,dataKey:i,index:p,className:"recharts-bar-background-rectangle"});return E.createElement(Kb,rf({key:"background-bar-".concat(p),option:r.props.background,isActive:p===s},m))})}},{key:"renderErrorBar",value:function(r,a){if(this.props.isAnimationActive&&!this.state.isAnimationFinished)return null;var o=this.props,i=o.data,s=o.xAxis,l=o.yAxis,u=o.layout,p=o.children,c=yn(p,Tu);if(!c)return null;var f=u==="vertical"?i[0].height/2:i[0].width/2,m=function(g,v){var y=Array.isArray(g.value)?g.value[1]:g.value;return{x:g.x,y:g.y,value:y,errorVal:It(g,v)}},h={clipPath:r?"url(#clipPath-".concat(a,")"):null};return E.createElement(Ne,h,c.map(function(d){return E.cloneElement(d,{key:"error-bar-".concat(a,"-").concat(d.props.dataKey),data:i,xAxis:s,yAxis:l,layout:u,offset:f,dataPointFormatter:m})}))}},{key:"render",value:function(){var r=this.props,a=r.hide,o=r.data,i=r.className,s=r.xAxis,l=r.yAxis,u=r.left,p=r.top,c=r.width,f=r.height,m=r.isAnimationActive,h=r.background,d=r.id;if(a||!o||!o.length)return null;var g=this.state.isAnimationFinished,v=ue("recharts-bar",i),y=s&&s.allowDataOverflow,x=l&&l.allowDataOverflow,P=y||x,w=le(d)?this.id:d;return E.createElement(Ne,{className:v},y||x?E.createElement("defs",null,E.createElement("clipPath",{id:"clipPath-".concat(w)},E.createElement("rect",{x:y?u:u-c/2,y:x?p:p-f/2,width:y?c:c*2,height:x?f:f*2}))):null,E.createElement(Ne,{className:"recharts-bar-rectangles",clipPath:P?"url(#clipPath-".concat(w,")"):null},h?this.renderBackground():null,this.renderRectangles()),this.renderErrorBar(P,w),(!m||g)&&Nr.renderCallByParent(this.props,o))}}],[{key:"getDerivedStateFromProps",value:function(r,a){return r.animationId!==a.prevAnimationId?{prevAnimationId:r.animationId,curData:r.data,prevData:a.curData}:r.data!==a.curData?{curData:r.data}:null}}])}(k.PureComponent);MP=Da;fa(Da,"displayName","Bar");fa(Da,"defaultProps",{xAxisId:0,yAxisId:0,legendType:"rect",minPointSize:0,hide:!1,data:[],layout:"vertical",activeBar:!1,isAnimationActive:!_o.isSsr,animationBegin:0,animationDuration:400,animationEasing:"ease"});fa(Da,"getComposedData",function(e){var t=e.props,n=e.item,r=e.barPosition,a=e.bandSize,o=e.xAxis,i=e.yAxis,s=e.xAxisTicks,l=e.yAxisTicks,u=e.stackedData,p=e.dataStartIndex,c=e.displayedData,f=e.offset,m=HX(r,n);if(!m)return null;var h=t.layout,d=n.type.defaultProps,g=d!==void 0?Ze(Ze({},d),n.props):n.props,v=g.dataKey,y=g.children,x=g.minPointSize,P=h==="horizontal"?i:o,w=u?P.scale.domain():null,S=YX({numericAxis:P}),O=yn(y,l7),C=c.map(function(_,T){var A,j,$,M,I,R;u?A=GX(u[p+T],w):(A=It(_,v),Array.isArray(A)||(A=[S,A]));var L=nte(x,MP.defaultProps.minPointSize)(A[1],T);if(h==="horizontal"){var N,D=[i.scale(A[0]),i.scale(A[1])],H=D[0],W=D[1];j=X3({axis:o,ticks:s,bandSize:a,offset:m.offset,entry:_,index:T}),$=(N=W??H)!==null&&N!==void 0?N:void 0,M=m.size;var G=H-W;if(I=Number.isNaN(G)?0:G,R={x:j,y:i.y,width:M,height:i.height},Math.abs(L)>0&&Math.abs(I)<Math.abs(L)){var Z=In(I||L)*(Math.abs(L)-Math.abs(I));$-=Z,I+=Z}}else{var re=[o.scale(A[0]),o.scale(A[1])],ve=re[0],be=re[1];if(j=ve,$=X3({axis:i,ticks:l,bandSize:a,offset:m.offset,entry:_,index:T}),M=be-ve,I=m.size,R={x:o.x,y:$,width:o.width,height:I},Math.abs(L)>0&&Math.abs(M)<Math.abs(L)){var J=In(M||L)*(Math.abs(L)-Math.abs(M));M+=J}}return Ze(Ze(Ze({},_),{},{x:j,y:$,width:M,height:I,value:u?A:A[1],payload:_,background:R},O&&O[T]&&O[T].props),{},{tooltipPayload:[mP(n,_)],tooltipPosition:{x:j+M/2,y:$+I/2}})});return Ze({data:C,layout:h},f)});function lu(e){"@babel/helpers - typeof";return lu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},lu(e)}function dte(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Qb(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,DP(r.key),r)}}function mte(e,t,n){return t&&Qb(e.prototype,t),n&&Qb(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function Zb(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Nn(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?Zb(Object(n),!0).forEach(function(r){bd(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):Zb(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function bd(e,t,n){return t=DP(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function DP(e){var t=hte(e,"string");return lu(t)=="symbol"?t:t+""}function hte(e,t){if(lu(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(lu(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var Eg=function(t,n,r,a,o){var i=t.width,s=t.height,l=t.layout,u=t.children,p=Object.keys(n),c={left:r.left,leftMirror:r.left,right:i-r.right,rightMirror:i-r.right,top:r.top,topMirror:r.top,bottom:s-r.bottom,bottomMirror:s-r.bottom},f=!!Yt(u,Da);return p.reduce(function(m,h){var d=n[h],g=d.orientation,v=d.domain,y=d.padding,x=y===void 0?{}:y,P=d.mirror,w=d.reversed,S="".concat(g).concat(P?"Mirror":""),O,C,_,T,A;if(d.type==="number"&&(d.padding==="gap"||d.padding==="no-gap")){var j=v[1]-v[0],$=1/0,M=d.categoricalDomain.sort(PM);if(M.forEach(function(re,ve){ve>0&&($=Math.min((re||0)-(M[ve-1]||0),$))}),Number.isFinite($)){var I=$/j,R=d.layout==="vertical"?r.height:r.width;if(d.padding==="gap"&&(O=I*R/2),d.padding==="no-gap"){var L=wo(t.barCategoryGap,I*R),N=I*R/2;O=N-L-(N-L)/R*L}}}a==="xAxis"?C=[r.left+(x.left||0)+(O||0),r.left+r.width-(x.right||0)-(O||0)]:a==="yAxis"?C=l==="horizontal"?[r.top+r.height-(x.bottom||0),r.top+(x.top||0)]:[r.top+(x.top||0)+(O||0),r.top+r.height-(x.bottom||0)-(O||0)]:C=d.range,w&&(C=[C[1],C[0]]);var D=BX(d,o,f),H=D.scale,W=D.realScaleType;H.domain(v).range(C),zX(H);var G=XX(H,Nn(Nn({},d),{},{realScaleType:W}));a==="xAxis"?(A=g==="top"&&!P||g==="bottom"&&P,_=r.left,T=c[S]-A*d.height):a==="yAxis"&&(A=g==="left"&&!P||g==="right"&&P,_=c[S]-A*d.width,T=r.top);var Z=Nn(Nn(Nn({},d),G),{},{realScaleType:W,x:_,y:T,scale:H,width:a==="xAxis"?r.width:d.width,height:a==="yAxis"?r.height:d.height});return Z.bandSize=Wp(Z,G),!d.hide&&a==="xAxis"?c[S]+=(A?-1:1)*Z.height:d.hide||(c[S]+=(A?-1:1)*Z.width),Nn(Nn({},m),{},bd({},h,Z))},{})},LP=function(t,n){var r=t.x,a=t.y,o=n.x,i=n.y;return{x:Math.min(r,o),y:Math.min(a,i),width:Math.abs(o-r),height:Math.abs(i-a)}},vte=function(t){var n=t.x1,r=t.y1,a=t.x2,o=t.y2;return LP({x:n,y:r},{x:a,y:o})},FP=function(){function e(t){dte(this,e),this.scale=t}return mte(e,[{key:"domain",get:function(){return this.scale.domain}},{key:"range",get:function(){return this.scale.range}},{key:"rangeMin",get:function(){return this.range()[0]}},{key:"rangeMax",get:function(){return this.range()[1]}},{key:"bandwidth",get:function(){return this.scale.bandwidth}},{key:"apply",value:function(n){var r=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{},a=r.bandAware,o=r.position;if(n!==void 0){if(o)switch(o){case"start":return this.scale(n);case"middle":{var i=this.bandwidth?this.bandwidth()/2:0;return this.scale(n)+i}case"end":{var s=this.bandwidth?this.bandwidth():0;return this.scale(n)+s}default:return this.scale(n)}if(a){var l=this.bandwidth?this.bandwidth()/2:0;return this.scale(n)+l}return this.scale(n)}}},{key:"isInRange",value:function(n){var r=this.range(),a=r[0],o=r[r.length-1];return a<=o?n>=a&&n<=o:n>=o&&n<=a}}],[{key:"create",value:function(n){return new e(n)}}])}();bd(FP,"EPS",1e-4);var Tg=function(t){var n=Object.keys(t).reduce(function(r,a){return Nn(Nn({},r),{},bd({},a,FP.create(t[a])))},{});return Nn(Nn({},n),{},{apply:function(a){var o=arguments.length>1&&arguments[1]!==void 0?arguments[1]:{},i=o.bandAware,s=o.position;return Iee(a,function(l,u){return n[u].apply(l,{bandAware:i,position:s})})},isInRange:function(a){return $P(a,function(o,i){return n[i].isInRange(o)})}})};function yte(e){return(e%180+180)%180}var gte=function(t){var n=t.width,r=t.height,a=arguments.length>1&&arguments[1]!==void 0?arguments[1]:0,o=yte(a),i=o*Math.PI/180,s=Math.atan(r/n),l=i>s&&i<Math.PI-s?r/Math.sin(i):n/Math.cos(i);return Math.abs(l)},xte=$a,wte=ku,bte=id;function Pte(e){return function(t,n,r){var a=Object(t);if(!wte(t)){var o=xte(n);t=bte(t),n=function(s){return o(a[s],s,a)}}var i=e(t,n,r);return i>-1?a[o?t[i]:i]:void 0}}var Ste=Pte,Ote=EP;function kte(e){var t=Ote(e),n=t%1;return t===t?n?t-n:t:0}var Cte=kte,_te=V8,Ate=$a,Ete=Cte,Tte=Math.max;function jte(e,t,n){var r=e==null?0:e.length;if(!r)return-1;var a=n==null?0:Ete(n);return a<0&&(a=Tte(r+a,0)),_te(e,Ate(t),a)}var Nte=jte,$te=Ste,Mte=Nte,Rte=$te(Mte),Ite=Rte;const Dte=_e(Ite);var Lte=k$(function(e){return{x:e.left,y:e.top,width:e.width,height:e.height}},function(e){return["l",e.left,"t",e.top,"w",e.width,"h",e.height].join("")}),jg=k.createContext(void 0),Ng=k.createContext(void 0),BP=k.createContext(void 0),zP=k.createContext({}),HP=k.createContext(void 0),GP=k.createContext(0),UP=k.createContext(0),Jb=function(t){var n=t.state,r=n.xAxisMap,a=n.yAxisMap,o=n.offset,i=t.clipPathId,s=t.children,l=t.width,u=t.height,p=Lte(o);return E.createElement(jg.Provider,{value:r},E.createElement(Ng.Provider,{value:a},E.createElement(zP.Provider,{value:o},E.createElement(BP.Provider,{value:p},E.createElement(HP.Provider,{value:i},E.createElement(GP.Provider,{value:u},E.createElement(UP.Provider,{value:l},s)))))))},Fte=function(){return k.useContext(HP)},WP=function(t){var n=k.useContext(jg);n==null&&Po();var r=n[t];return r==null&&Po(),r},Bte=function(){var t=k.useContext(jg);return ra(t)},zte=function(){var t=k.useContext(Ng),n=Dte(t,function(r){return $P(r.domain,Number.isFinite)});return n||ra(t)},qP=function(t){var n=k.useContext(Ng);n==null&&Po();var r=n[t];return r==null&&Po(),r},Hte=function(){var t=k.useContext(BP);return t},Gte=function(){return k.useContext(zP)},$g=function(){return k.useContext(UP)},Mg=function(){return k.useContext(GP)};function Wi(e){"@babel/helpers - typeof";return Wi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Wi(e)}function Ute(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Wte(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,KP(r.key),r)}}function qte(e,t,n){return t&&Wte(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function Vte(e,t,n){return t=of(t),Kte(e,VP()?Reflect.construct(t,n||[],of(e).constructor):t.apply(e,n))}function Kte(e,t){if(t&&(Wi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return Xte(e)}function Xte(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function VP(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(VP=function(){return!!e})()}function of(e){return of=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},of(e)}function Yte(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&tv(e,t)}function tv(e,t){return tv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},tv(e,t)}function e2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function t2(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?e2(Object(n),!0).forEach(function(r){Rg(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):e2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function Rg(e,t,n){return t=KP(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function KP(e){var t=Qte(e,"string");return Wi(t)=="symbol"?t:t+""}function Qte(e,t){if(Wi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Wi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}function Zte(e,t){return nne(e)||tne(e,t)||ene(e,t)||Jte()}function Jte(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function ene(e,t){if(e){if(typeof e=="string")return n2(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return n2(e,t)}}function n2(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function tne(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function nne(e){if(Array.isArray(e))return e}function nv(){return nv=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},nv.apply(this,arguments)}var rne=function(t,n){var r;return E.isValidElement(t)?r=E.cloneElement(t,n):ae(t)?r=t(n):r=E.createElement("line",nv({},n,{className:"recharts-reference-line-line"})),r},ane=function(t,n,r,a,o,i,s,l,u){var p=o.x,c=o.y,f=o.width,m=o.height;if(r){var h=u.y,d=t.y.apply(h,{position:i});if(pr(u,"discard")&&!t.y.isInRange(d))return null;var g=[{x:p+f,y:d},{x:p,y:d}];return l==="left"?g.reverse():g}if(n){var v=u.x,y=t.x.apply(v,{position:i});if(pr(u,"discard")&&!t.x.isInRange(y))return null;var x=[{x:y,y:c+m},{x:y,y:c}];return s==="top"?x.reverse():x}if(a){var P=u.segment,w=P.map(function(S){return t.apply(S,{position:i})});return pr(u,"discard")&&Aee(w,function(S){return!t.isInRange(S)})?null:w}return null};function one(e){var t=e.x,n=e.y,r=e.segment,a=e.xAxisId,o=e.yAxisId,i=e.shape,s=e.className,l=e.alwaysShow,u=Fte(),p=WP(a),c=qP(o),f=Hte();if(!u||!f)return null;Tr(l===void 0,'The alwaysShow prop is deprecated. Please use ifOverflow="extendDomain" instead.');var m=Tg({x:p.scale,y:c.scale}),h=ot(t),d=ot(n),g=r&&r.length===2,v=ane(m,h,d,g,f,e.position,p.orientation,c.orientation,e);if(!v)return null;var y=Zte(v,2),x=y[0],P=x.x,w=x.y,S=y[1],O=S.x,C=S.y,_=pr(e,"hidden")?"url(#".concat(u,")"):void 0,T=t2(t2({clipPath:_},ie(e,!0)),{},{x1:P,y1:w,x2:O,y2:C});return E.createElement(Ne,{className:ue("recharts-reference-line",s)},rne(i,T),kt.renderCallByParent(e,vte({x1:P,y1:w,x2:O,y2:C})))}var Ig=function(e){function t(){return Ute(this,t),Vte(this,t,arguments)}return Yte(t,e),qte(t,[{key:"render",value:function(){return E.createElement(one,this.props)}}])}(E.Component);Rg(Ig,"displayName","ReferenceLine");Rg(Ig,"defaultProps",{isFront:!1,ifOverflow:"discard",xAxisId:0,yAxisId:0,fill:"none",stroke:"#ccc",fillOpacity:1,strokeWidth:1,position:"middle"});function rv(){return rv=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},rv.apply(this,arguments)}function qi(e){"@babel/helpers - typeof";return qi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},qi(e)}function r2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function a2(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?r2(Object(n),!0).forEach(function(r){Pd(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):r2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function ine(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function sne(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,YP(r.key),r)}}function lne(e,t,n){return t&&sne(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function une(e,t,n){return t=sf(t),cne(e,XP()?Reflect.construct(t,n||[],sf(e).constructor):t.apply(e,n))}function cne(e,t){if(t&&(qi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return pne(e)}function pne(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function XP(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(XP=function(){return!!e})()}function sf(e){return sf=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},sf(e)}function fne(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&av(e,t)}function av(e,t){return av=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},av(e,t)}function Pd(e,t,n){return t=YP(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function YP(e){var t=dne(e,"string");return qi(t)=="symbol"?t:t+""}function dne(e,t){if(qi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(qi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var mne=function(t){var n=t.x,r=t.y,a=t.xAxis,o=t.yAxis,i=Tg({x:a.scale,y:o.scale}),s=i.apply({x:n,y:r},{bandAware:!0});return pr(t,"discard")&&!i.isInRange(s)?null:s},Sd=function(e){function t(){return ine(this,t),une(this,t,arguments)}return fne(t,e),lne(t,[{key:"render",value:function(){var r=this.props,a=r.x,o=r.y,i=r.r,s=r.alwaysShow,l=r.clipPathId,u=ot(a),p=ot(o);if(Tr(s===void 0,'The alwaysShow prop is deprecated. Please use ifOverflow="extendDomain" instead.'),!u||!p)return null;var c=mne(this.props);if(!c)return null;var f=c.x,m=c.y,h=this.props,d=h.shape,g=h.className,v=pr(this.props,"hidden")?"url(#".concat(l,")"):void 0,y=a2(a2({clipPath:v},ie(this.props,!0)),{},{cx:f,cy:m});return E.createElement(Ne,{className:ue("recharts-reference-dot",g)},t.renderDot(d,y),kt.renderCallByParent(this.props,{x:f-i,y:m-i,width:2*i,height:2*i}))}}])}(E.Component);Pd(Sd,"displayName","ReferenceDot");Pd(Sd,"defaultProps",{isFront:!1,ifOverflow:"discard",xAxisId:0,yAxisId:0,r:10,fill:"#fff",stroke:"#ccc",fillOpacity:1,strokeWidth:1});Pd(Sd,"renderDot",function(e,t){var n;return E.isValidElement(e)?n=E.cloneElement(e,t):ae(e)?n=e(t):n=E.createElement(gd,rv({},t,{cx:t.cx,cy:t.cy,className:"recharts-reference-dot-dot"})),n});function ov(){return ov=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},ov.apply(this,arguments)}function Vi(e){"@babel/helpers - typeof";return Vi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Vi(e)}function o2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function i2(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?o2(Object(n),!0).forEach(function(r){Od(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):o2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function hne(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function vne(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,ZP(r.key),r)}}function yne(e,t,n){return t&&vne(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function gne(e,t,n){return t=lf(t),xne(e,QP()?Reflect.construct(t,n||[],lf(e).constructor):t.apply(e,n))}function xne(e,t){if(t&&(Vi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return wne(e)}function wne(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function QP(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(QP=function(){return!!e})()}function lf(e){return lf=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},lf(e)}function bne(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&iv(e,t)}function iv(e,t){return iv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},iv(e,t)}function Od(e,t,n){return t=ZP(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function ZP(e){var t=Pne(e,"string");return Vi(t)=="symbol"?t:t+""}function Pne(e,t){if(Vi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Vi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var Sne=function(t,n,r,a,o){var i=o.x1,s=o.x2,l=o.y1,u=o.y2,p=o.xAxis,c=o.yAxis;if(!p||!c)return null;var f=Tg({x:p.scale,y:c.scale}),m={x:t?f.x.apply(i,{position:"start"}):f.x.rangeMin,y:r?f.y.apply(l,{position:"start"}):f.y.rangeMin},h={x:n?f.x.apply(s,{position:"end"}):f.x.rangeMax,y:a?f.y.apply(u,{position:"end"}):f.y.rangeMax};return pr(o,"discard")&&(!f.isInRange(m)||!f.isInRange(h))?null:LP(m,h)},kd=function(e){function t(){return hne(this,t),gne(this,t,arguments)}return bne(t,e),yne(t,[{key:"render",value:function(){var r=this.props,a=r.x1,o=r.x2,i=r.y1,s=r.y2,l=r.className,u=r.alwaysShow,p=r.clipPathId;Tr(u===void 0,'The alwaysShow prop is deprecated. Please use ifOverflow="extendDomain" instead.');var c=ot(a),f=ot(o),m=ot(i),h=ot(s),d=this.props.shape;if(!c&&!f&&!m&&!h&&!d)return null;var g=Sne(c,f,m,h,this.props);if(!g&&!d)return null;var v=pr(this.props,"hidden")?"url(#".concat(p,")"):void 0;return E.createElement(Ne,{className:ue("recharts-reference-area",l)},t.renderRect(d,i2(i2({clipPath:v},ie(this.props,!0)),g)),kt.renderCallByParent(this.props,g))}}])}(E.Component);Od(kd,"displayName","ReferenceArea");Od(kd,"defaultProps",{isFront:!1,ifOverflow:"discard",xAxisId:0,yAxisId:0,r:10,fill:"#ccc",fillOpacity:.5,stroke:"none",strokeWidth:1});Od(kd,"renderRect",function(e,t){var n;return E.isValidElement(e)?n=E.cloneElement(e,t):ae(e)?n=e(t):n=E.createElement(Ag,ov({},t,{className:"recharts-reference-area-rect"})),n});function JP(e,t,n){if(t<1)return[];if(t===1&&n===void 0)return e;for(var r=[],a=0;a<e.length;a+=t)r.push(e[a]);return r}function One(e,t,n){var r={width:e.width+t.width,height:e.height+t.height};return gte(r,n)}function kne(e,t,n){var r=n==="width",a=e.x,o=e.y,i=e.width,s=e.height;return t===1?{start:r?a:o,end:r?a+i:o+s}:{start:r?a+i:o+s,end:r?a:o}}function uf(e,t,n,r,a){if(e*t<e*r||e*t>e*a)return!1;var o=n();return e*(t-e*o/2-r)>=0&&e*(t+e*o/2-a)<=0}function Cne(e,t){return JP(e,t+1)}function _ne(e,t,n,r,a){for(var o=(r||[]).slice(),i=t.start,s=t.end,l=0,u=1,p=i,c=function(){var h=r==null?void 0:r[l];if(h===void 0)return{v:JP(r,u)};var d=l,g,v=function(){return g===void 0&&(g=n(h,d)),g},y=h.coordinate,x=l===0||uf(e,y,v,p,s);x||(l=0,p=i,u+=1),x&&(p=y+e*(v()/2+a),l+=u)},f;u<=o.length;)if(f=c(),f)return f.v;return[]}function uu(e){"@babel/helpers - typeof";return uu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},uu(e)}function s2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function St(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?s2(Object(n),!0).forEach(function(r){Ane(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):s2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function Ane(e,t,n){return t=Ene(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function Ene(e){var t=Tne(e,"string");return uu(t)=="symbol"?t:t+""}function Tne(e,t){if(uu(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(uu(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function jne(e,t,n,r,a){for(var o=(r||[]).slice(),i=o.length,s=t.start,l=t.end,u=function(f){var m=o[f],h,d=function(){return h===void 0&&(h=n(m,f)),h};if(f===i-1){var g=e*(m.coordinate+e*d()/2-l);o[f]=m=St(St({},m),{},{tickCoord:g>0?m.coordinate-g*e:m.coordinate})}else o[f]=m=St(St({},m),{},{tickCoord:m.coordinate});var v=uf(e,m.tickCoord,d,s,l);v&&(l=m.tickCoord-e*(d()/2+a),o[f]=St(St({},m),{},{isShow:!0}))},p=i-1;p>=0;p--)u(p);return o}function Nne(e,t,n,r,a,o){var i=(r||[]).slice(),s=i.length,l=t.start,u=t.end;if(o){var p=r[s-1],c=n(p,s-1),f=e*(p.coordinate+e*c/2-u);i[s-1]=p=St(St({},p),{},{tickCoord:f>0?p.coordinate-f*e:p.coordinate});var m=uf(e,p.tickCoord,function(){return c},l,u);m&&(u=p.tickCoord-e*(c/2+a),i[s-1]=St(St({},p),{},{isShow:!0}))}for(var h=o?s-1:s,d=function(y){var x=i[y],P,w=function(){return P===void 0&&(P=n(x,y)),P};if(y===0){var S=e*(x.coordinate-e*w()/2-l);i[y]=x=St(St({},x),{},{tickCoord:S<0?x.coordinate-S*e:x.coordinate})}else i[y]=x=St(St({},x),{},{tickCoord:x.coordinate});var O=uf(e,x.tickCoord,w,l,u);O&&(l=x.tickCoord+e*(w()/2+a),i[y]=St(St({},x),{},{isShow:!0}))},g=0;g<h;g++)d(g);return i}function Dg(e,t,n){var r=e.tick,a=e.ticks,o=e.viewBox,i=e.minTickGap,s=e.orientation,l=e.interval,u=e.tickFormatter,p=e.unit,c=e.angle;if(!a||!a.length||!r)return[];if(V(l)||_o.isSsr)return Cne(a,typeof l=="number"&&V(l)?l:0);var f=[],m=s==="top"||s==="bottom"?"width":"height",h=p&&m==="width"?il(p,{fontSize:t,letterSpacing:n}):{width:0,height:0},d=function(x,P){var w=ae(u)?u(x.value,P):x.value;return m==="width"?One(il(w,{fontSize:t,letterSpacing:n}),h,c):il(w,{fontSize:t,letterSpacing:n})[m]},g=a.length>=2?In(a[1].coordinate-a[0].coordinate):1,v=kne(o,g,m);return l==="equidistantPreserveStart"?_ne(g,v,d,a,i):(l==="preserveStart"||l==="preserveStartEnd"?f=Nne(g,v,d,a,i,l==="preserveStartEnd"):f=jne(g,v,d,a,i),f.filter(function(y){return y.isShow}))}var $ne=["viewBox"],Mne=["viewBox"],Rne=["ticks"];function Ki(e){"@babel/helpers - typeof";return Ki=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Ki(e)}function Zo(){return Zo=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Zo.apply(this,arguments)}function l2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function tt(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?l2(Object(n),!0).forEach(function(r){Lg(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):l2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function Bm(e,t){if(e==null)return{};var n=Ine(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function Ine(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function Dne(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function u2(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,tS(r.key),r)}}function Lne(e,t,n){return t&&u2(e.prototype,t),n&&u2(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function Fne(e,t,n){return t=cf(t),Bne(e,eS()?Reflect.construct(t,n||[],cf(e).constructor):t.apply(e,n))}function Bne(e,t){if(t&&(Ki(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return zne(e)}function zne(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function eS(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(eS=function(){return!!e})()}function cf(e){return cf=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},cf(e)}function Hne(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&sv(e,t)}function sv(e,t){return sv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},sv(e,t)}function Lg(e,t,n){return t=tS(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function tS(e){var t=Gne(e,"string");return Ki(t)=="symbol"?t:t+""}function Gne(e,t){if(Ki(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Ki(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var Os=function(e){function t(n){var r;return Dne(this,t),r=Fne(this,t,[n]),r.state={fontSize:"",letterSpacing:""},r}return Hne(t,e),Lne(t,[{key:"shouldComponentUpdate",value:function(r,a){var o=r.viewBox,i=Bm(r,$ne),s=this.props,l=s.viewBox,u=Bm(s,Mne);return!ui(o,l)||!ui(i,u)||!ui(a,this.state)}},{key:"componentDidMount",value:function(){var r=this.layerReference;if(r){var a=r.getElementsByClassName("recharts-cartesian-axis-tick-value")[0];a&&this.setState({fontSize:window.getComputedStyle(a).fontSize,letterSpacing:window.getComputedStyle(a).letterSpacing})}}},{key:"getTickLineCoord",value:function(r){var a=this.props,o=a.x,i=a.y,s=a.width,l=a.height,u=a.orientation,p=a.tickSize,c=a.mirror,f=a.tickMargin,m,h,d,g,v,y,x=c?-1:1,P=r.tickSize||p,w=V(r.tickCoord)?r.tickCoord:r.coordinate;switch(u){case"top":m=h=r.coordinate,g=i+ +!c*l,d=g-x*P,y=d-x*f,v=w;break;case"left":d=g=r.coordinate,h=o+ +!c*s,m=h-x*P,v=m-x*f,y=w;break;case"right":d=g=r.coordinate,h=o+ +c*s,m=h+x*P,v=m+x*f,y=w;break;default:m=h=r.coordinate,g=i+ +c*l,d=g+x*P,y=d+x*f,v=w;break}return{line:{x1:m,y1:d,x2:h,y2:g},tick:{x:v,y}}}},{key:"getTickTextAnchor",value:function(){var r=this.props,a=r.orientation,o=r.mirror,i;switch(a){case"left":i=o?"start":"end";break;case"right":i=o?"end":"start";break;default:i="middle";break}return i}},{key:"getTickVerticalAnchor",value:function(){var r=this.props,a=r.orientation,o=r.mirror,i="end";switch(a){case"left":case"right":i="middle";break;case"top":i=o?"start":"end";break;default:i=o?"end":"start";break}return i}},{key:"renderAxisLine",value:function(){var r=this.props,a=r.x,o=r.y,i=r.width,s=r.height,l=r.orientation,u=r.mirror,p=r.axisLine,c=tt(tt(tt({},ie(this.props,!1)),ie(p,!1)),{},{fill:"none"});if(l==="top"||l==="bottom"){var f=+(l==="top"&&!u||l==="bottom"&&u);c=tt(tt({},c),{},{x1:a,y1:o+f*s,x2:a+i,y2:o+f*s})}else{var m=+(l==="left"&&!u||l==="right"&&u);c=tt(tt({},c),{},{x1:a+m*i,y1:o,x2:a+m*i,y2:o+s})}return E.createElement("line",Zo({},c,{className:ue("recharts-cartesian-axis-line",vn(p,"className"))}))}},{key:"renderTicks",value:function(r,a,o){var i=this,s=this.props,l=s.tickLine,u=s.stroke,p=s.tick,c=s.tickFormatter,f=s.unit,m=Dg(tt(tt({},this.props),{},{ticks:r}),a,o),h=this.getTickTextAnchor(),d=this.getTickVerticalAnchor(),g=ie(this.props,!1),v=ie(p,!1),y=tt(tt({},g),{},{fill:"none"},ie(l,!1)),x=m.map(function(P,w){var S=i.getTickLineCoord(P),O=S.line,C=S.tick,_=tt(tt(tt(tt({textAnchor:h,verticalAnchor:d},g),{},{stroke:"none",fill:u},v),C),{},{index:w,payload:P,visibleTicksCount:m.length,tickFormatter:c});return E.createElement(Ne,Zo({className:"recharts-cartesian-axis-tick",key:"tick-".concat(P.value,"-").concat(P.coordinate,"-").concat(P.tickCoord)},vp(i.props,P,w)),l&&E.createElement("line",Zo({},y,O,{className:ue("recharts-cartesian-axis-tick-line",vn(l,"className"))})),p&&t.renderTickItem(p,_,"".concat(ae(c)?c(P.value,w):P.value).concat(f||"")))});return E.createElement("g",{className:"recharts-cartesian-axis-ticks"},x)}},{key:"render",value:function(){var r=this,a=this.props,o=a.axisLine,i=a.width,s=a.height,l=a.ticksGenerator,u=a.className,p=a.hide;if(p)return null;var c=this.props,f=c.ticks,m=Bm(c,Rne),h=f;return ae(l)&&(h=f&&f.length>0?l(this.props):l(m)),i<=0||s<=0||!h||!h.length?null:E.createElement(Ne,{className:ue("recharts-cartesian-axis",u),ref:function(g){r.layerReference=g}},o&&this.renderAxisLine(),this.renderTicks(h,this.state.fontSize,this.state.letterSpacing),kt.renderCallByParent(this.props))}}],[{key:"renderTickItem",value:function(r,a,o){var i,s=ue(a.className,"recharts-cartesian-axis-tick-value");return E.isValidElement(r)?i=E.cloneElement(r,tt(tt({},a),{},{className:s})):ae(r)?i=r(tt(tt({},a),{},{className:s})):i=E.createElement(Tp,Zo({},a,{className:"recharts-cartesian-axis-tick-value"}),o),i}}])}(k.Component);Lg(Os,"displayName","CartesianAxis");Lg(Os,"defaultProps",{x:0,y:0,width:0,height:0,viewBox:{x:0,y:0,width:0,height:0},orientation:"bottom",ticks:[],stroke:"#666",tickLine:!0,axisLine:!0,tick:!0,mirror:!1,minTickGap:5,tickSize:6,tickMargin:2,interval:"preserveEnd"});var Une=["x1","y1","x2","y2","key"],Wne=["offset"];function So(e){"@babel/helpers - typeof";return So=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},So(e)}function c2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Ct(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?c2(Object(n),!0).forEach(function(r){qne(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):c2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function qne(e,t,n){return t=Vne(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function Vne(e){var t=Kne(e,"string");return So(t)=="symbol"?t:t+""}function Kne(e,t){if(So(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(So(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function Za(){return Za=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Za.apply(this,arguments)}function p2(e,t){if(e==null)return{};var n=Xne(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function Xne(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}var Yne=function(t){var n=t.fill;if(!n||n==="none")return null;var r=t.fillOpacity,a=t.x,o=t.y,i=t.width,s=t.height,l=t.ry;return E.createElement("rect",{x:a,y:o,ry:l,width:i,height:s,stroke:"none",fill:n,fillOpacity:r,className:"recharts-cartesian-grid-bg"})};function nS(e,t){var n;if(E.isValidElement(e))n=E.cloneElement(e,t);else if(ae(e))n=e(t);else{var r=t.x1,a=t.y1,o=t.x2,i=t.y2,s=t.key,l=p2(t,Une),u=ie(l,!1);u.offset;var p=p2(u,Wne);n=E.createElement("line",Za({},p,{x1:r,y1:a,x2:o,y2:i,fill:"none",key:s}))}return n}function Qne(e){var t=e.x,n=e.width,r=e.horizontal,a=r===void 0?!0:r,o=e.horizontalPoints;if(!a||!o||!o.length)return null;var i=o.map(function(s,l){var u=Ct(Ct({},e),{},{x1:t,y1:s,x2:t+n,y2:s,key:"line-".concat(l),index:l});return nS(a,u)});return E.createElement("g",{className:"recharts-cartesian-grid-horizontal"},i)}function Zne(e){var t=e.y,n=e.height,r=e.vertical,a=r===void 0?!0:r,o=e.verticalPoints;if(!a||!o||!o.length)return null;var i=o.map(function(s,l){var u=Ct(Ct({},e),{},{x1:s,y1:t,x2:s,y2:t+n,key:"line-".concat(l),index:l});return nS(a,u)});return E.createElement("g",{className:"recharts-cartesian-grid-vertical"},i)}function Jne(e){var t=e.horizontalFill,n=e.fillOpacity,r=e.x,a=e.y,o=e.width,i=e.height,s=e.horizontalPoints,l=e.horizontal,u=l===void 0?!0:l;if(!u||!t||!t.length)return null;var p=s.map(function(f){return Math.round(f+a-a)}).sort(function(f,m){return f-m});a!==p[0]&&p.unshift(0);var c=p.map(function(f,m){var h=!p[m+1],d=h?a+i-f:p[m+1]-f;if(d<=0)return null;var g=m%t.length;return E.createElement("rect",{key:"react-".concat(m),y:f,x:r,height:d,width:o,stroke:"none",fill:t[g],fillOpacity:n,className:"recharts-cartesian-grid-bg"})});return E.createElement("g",{className:"recharts-cartesian-gridstripes-horizontal"},c)}function ere(e){var t=e.vertical,n=t===void 0?!0:t,r=e.verticalFill,a=e.fillOpacity,o=e.x,i=e.y,s=e.width,l=e.height,u=e.verticalPoints;if(!n||!r||!r.length)return null;var p=u.map(function(f){return Math.round(f+o-o)}).sort(function(f,m){return f-m});o!==p[0]&&p.unshift(0);var c=p.map(function(f,m){var h=!p[m+1],d=h?o+s-f:p[m+1]-f;if(d<=0)return null;var g=m%r.length;return E.createElement("rect",{key:"react-".concat(m),x:f,y:i,width:d,height:l,stroke:"none",fill:r[g],fillOpacity:a,className:"recharts-cartesian-grid-bg"})});return E.createElement("g",{className:"recharts-cartesian-gridstripes-vertical"},c)}var tre=function(t,n){var r=t.xAxis,a=t.width,o=t.height,i=t.offset;return fP(Dg(Ct(Ct(Ct({},Os.defaultProps),r),{},{ticks:_r(r,!0),viewBox:{x:0,y:0,width:a,height:o}})),i.left,i.left+i.width,n)},nre=function(t,n){var r=t.yAxis,a=t.width,o=t.height,i=t.offset;return fP(Dg(Ct(Ct(Ct({},Os.defaultProps),r),{},{ticks:_r(r,!0),viewBox:{x:0,y:0,width:a,height:o}})),i.top,i.top+i.height,n)},Do={horizontal:!0,vertical:!0,stroke:"#ccc",fill:"none",verticalFill:[],horizontalFill:[]};function Dn(e){var t,n,r,a,o,i,s=$g(),l=Mg(),u=Gte(),p=Ct(Ct({},e),{},{stroke:(t=e.stroke)!==null&&t!==void 0?t:Do.stroke,fill:(n=e.fill)!==null&&n!==void 0?n:Do.fill,horizontal:(r=e.horizontal)!==null&&r!==void 0?r:Do.horizontal,horizontalFill:(a=e.horizontalFill)!==null&&a!==void 0?a:Do.horizontalFill,vertical:(o=e.vertical)!==null&&o!==void 0?o:Do.vertical,verticalFill:(i=e.verticalFill)!==null&&i!==void 0?i:Do.verticalFill,x:V(e.x)?e.x:u.left,y:V(e.y)?e.y:u.top,width:V(e.width)?e.width:u.width,height:V(e.height)?e.height:u.height}),c=p.x,f=p.y,m=p.width,h=p.height,d=p.syncWithTicks,g=p.horizontalValues,v=p.verticalValues,y=Bte(),x=zte();if(!V(m)||m<=0||!V(h)||h<=0||!V(c)||c!==+c||!V(f)||f!==+f)return null;var P=p.verticalCoordinatesGenerator||tre,w=p.horizontalCoordinatesGenerator||nre,S=p.horizontalPoints,O=p.verticalPoints;if((!S||!S.length)&&ae(w)){var C=g&&g.length,_=w({yAxis:x?Ct(Ct({},x),{},{ticks:C?g:x.ticks}):void 0,width:s,height:l,offset:u},C?!0:d);Tr(Array.isArray(_),"horizontalCoordinatesGenerator should return Array but instead it returned [".concat(So(_),"]")),Array.isArray(_)&&(S=_)}if((!O||!O.length)&&ae(P)){var T=v&&v.length,A=P({xAxis:y?Ct(Ct({},y),{},{ticks:T?v:y.ticks}):void 0,width:s,height:l,offset:u},T?!0:d);Tr(Array.isArray(A),"verticalCoordinatesGenerator should return Array but instead it returned [".concat(So(A),"]")),Array.isArray(A)&&(O=A)}return E.createElement("g",{className:"recharts-cartesian-grid"},E.createElement(Yne,{fill:p.fill,fillOpacity:p.fillOpacity,x:p.x,y:p.y,width:p.width,height:p.height,ry:p.ry}),E.createElement(Qne,Za({},p,{offset:u,horizontalPoints:S,xAxis:y,yAxis:x})),E.createElement(Zne,Za({},p,{offset:u,verticalPoints:O,xAxis:y,yAxis:x})),E.createElement(Jne,Za({},p,{horizontalPoints:S})),E.createElement(ere,Za({},p,{verticalPoints:O})))}Dn.displayName="CartesianGrid";var rre=["type","layout","connectNulls","ref"],are=["key"];function Xi(e){"@babel/helpers - typeof";return Xi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Xi(e)}function f2(e,t){if(e==null)return{};var n=ore(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function ore(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function cl(){return cl=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},cl.apply(this,arguments)}function d2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Vt(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?d2(Object(n),!0).forEach(function(r){$n(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):d2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function Lo(e){return ure(e)||lre(e)||sre(e)||ire()}function ire(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function sre(e,t){if(e){if(typeof e=="string")return lv(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return lv(e,t)}}function lre(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function ure(e){if(Array.isArray(e))return lv(e)}function lv(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function cre(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function m2(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,aS(r.key),r)}}function pre(e,t,n){return t&&m2(e.prototype,t),n&&m2(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function fre(e,t,n){return t=pf(t),dre(e,rS()?Reflect.construct(t,n||[],pf(e).constructor):t.apply(e,n))}function dre(e,t){if(t&&(Xi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return mre(e)}function mre(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function rS(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(rS=function(){return!!e})()}function pf(e){return pf=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},pf(e)}function hre(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&uv(e,t)}function uv(e,t){return uv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},uv(e,t)}function $n(e,t,n){return t=aS(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function aS(e){var t=vre(e,"string");return Xi(t)=="symbol"?t:t+""}function vre(e,t){if(Xi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Xi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var Bt=function(e){function t(){var n;cre(this,t);for(var r=arguments.length,a=new Array(r),o=0;o<r;o++)a[o]=arguments[o];return n=fre(this,t,[].concat(a)),$n(n,"state",{isAnimationFinished:!0,totalLength:0}),$n(n,"generateSimpleStrokeDasharray",function(i,s){return"".concat(s,"px ").concat(i-s,"px")}),$n(n,"getStrokeDasharray",function(i,s,l){var u=l.reduce(function(v,y){return v+y});if(!u)return n.generateSimpleStrokeDasharray(s,i);for(var p=Math.floor(i/u),c=i%u,f=s-i,m=[],h=0,d=0;h<l.length;d+=l[h],++h)if(d+l[h]>c){m=[].concat(Lo(l.slice(0,h)),[c-d]);break}var g=m.length%2===0?[0,f]:[f];return[].concat(Lo(t.repeat(l,p)),Lo(m),g).map(function(v){return"".concat(v,"px")}).join(", ")}),$n(n,"id",ys("recharts-line-")),$n(n,"pathRef",function(i){n.mainCurve=i}),$n(n,"handleAnimationEnd",function(){n.setState({isAnimationFinished:!0}),n.props.onAnimationEnd&&n.props.onAnimationEnd()}),$n(n,"handleAnimationStart",function(){n.setState({isAnimationFinished:!1}),n.props.onAnimationStart&&n.props.onAnimationStart()}),n}return hre(t,e),pre(t,[{key:"componentDidMount",value:function(){if(this.props.isAnimationActive){var r=this.getTotalLength();this.setState({totalLength:r})}}},{key:"componentDidUpdate",value:function(){if(this.props.isAnimationActive){var r=this.getTotalLength();r!==this.state.totalLength&&this.setState({totalLength:r})}}},{key:"getTotalLength",value:function(){var r=this.mainCurve;try{return r&&r.getTotalLength&&r.getTotalLength()||0}catch{return 0}}},{key:"renderErrorBar",value:function(r,a){if(this.props.isAnimationActive&&!this.state.isAnimationFinished)return null;var o=this.props,i=o.points,s=o.xAxis,l=o.yAxis,u=o.layout,p=o.children,c=yn(p,Tu);if(!c)return null;var f=function(d,g){return{x:d.x,y:d.y,value:d.value,errorVal:It(d.payload,g)}},m={clipPath:r?"url(#clipPath-".concat(a,")"):null};return E.createElement(Ne,m,c.map(function(h){return E.cloneElement(h,{key:"bar-".concat(h.props.dataKey),data:i,xAxis:s,yAxis:l,layout:u,dataPointFormatter:f})}))}},{key:"renderDots",value:function(r,a,o){var i=this.props.isAnimationActive;if(i&&!this.state.isAnimationFinished)return null;var s=this.props,l=s.dot,u=s.points,p=s.dataKey,c=ie(this.props,!1),f=ie(l,!0),m=u.map(function(d,g){var v=Vt(Vt(Vt({key:"dot-".concat(g),r:3},c),f),{},{index:g,cx:d.x,cy:d.y,value:d.value,dataKey:p,payload:d.payload,points:u});return t.renderDotItem(l,v)}),h={clipPath:r?"url(#clipPath-".concat(a?"":"dots-").concat(o,")"):null};return E.createElement(Ne,cl({className:"recharts-line-dots",key:"dots"},h),m)}},{key:"renderCurveStatically",value:function(r,a,o,i){var s=this.props,l=s.type,u=s.layout,p=s.connectNulls;s.ref;var c=f2(s,rre),f=Vt(Vt(Vt({},ie(c,!0)),{},{fill:"none",className:"recharts-line-curve",clipPath:a?"url(#clipPath-".concat(o,")"):null,points:r},i),{},{type:l,layout:u,connectNulls:p});return E.createElement(di,cl({},f,{pathRef:this.pathRef}))}},{key:"renderCurveWithAnimation",value:function(r,a){var o=this,i=this.props,s=i.points,l=i.strokeDasharray,u=i.isAnimationActive,p=i.animationBegin,c=i.animationDuration,f=i.animationEasing,m=i.animationId,h=i.animateNewValues,d=i.width,g=i.height,v=this.state,y=v.prevPoints,x=v.totalLength;return E.createElement(mr,{begin:p,duration:c,isActive:u,easing:f,from:{t:0},to:{t:1},key:"line-".concat(m),onAnimationEnd:this.handleAnimationEnd,onAnimationStart:this.handleAnimationStart},function(P){var w=P.t;if(y){var S=y.length/s.length,O=s.map(function(j,$){var M=Math.floor($*S);if(y[M]){var I=y[M],R=dt(I.x,j.x),L=dt(I.y,j.y);return Vt(Vt({},j),{},{x:R(w),y:L(w)})}if(h){var N=dt(d*2,j.x),D=dt(g/2,j.y);return Vt(Vt({},j),{},{x:N(w),y:D(w)})}return Vt(Vt({},j),{},{x:j.x,y:j.y})});return o.renderCurveStatically(O,r,a)}var C=dt(0,x),_=C(w),T;if(l){var A="".concat(l).split(/[,\s]+/gim).map(function(j){return parseFloat(j)});T=o.getStrokeDasharray(_,x,A)}else T=o.generateSimpleStrokeDasharray(x,_);return o.renderCurveStatically(s,r,a,{strokeDasharray:T})})}},{key:"renderCurve",value:function(r,a){var o=this.props,i=o.points,s=o.isAnimationActive,l=this.state,u=l.prevPoints,p=l.totalLength;return s&&i&&i.length&&(!u&&p>0||!Fi(u,i))?this.renderCurveWithAnimation(r,a):this.renderCurveStatically(i,r,a)}},{key:"render",value:function(){var r,a=this.props,o=a.hide,i=a.dot,s=a.points,l=a.className,u=a.xAxis,p=a.yAxis,c=a.top,f=a.left,m=a.width,h=a.height,d=a.isAnimationActive,g=a.id;if(o||!s||!s.length)return null;var v=this.state.isAnimationFinished,y=s.length===1,x=ue("recharts-line",l),P=u&&u.allowDataOverflow,w=p&&p.allowDataOverflow,S=P||w,O=le(g)?this.id:g,C=(r=ie(i,!1))!==null&&r!==void 0?r:{r:3,strokeWidth:2},_=C.r,T=_===void 0?3:_,A=C.strokeWidth,j=A===void 0?2:A,$=a8(i)?i:{},M=$.clipDot,I=M===void 0?!0:M,R=T*2+j;return E.createElement(Ne,{className:x},P||w?E.createElement("defs",null,E.createElement("clipPath",{id:"clipPath-".concat(O)},E.createElement("rect",{x:P?f:f-m/2,y:w?c:c-h/2,width:P?m:m*2,height:w?h:h*2})),!I&&E.createElement("clipPath",{id:"clipPath-dots-".concat(O)},E.createElement("rect",{x:f-R/2,y:c-R/2,width:m+R,height:h+R}))):null,!y&&this.renderCurve(S,O),this.renderErrorBar(S,O),(y||i)&&this.renderDots(S,I,O),(!d||v)&&Nr.renderCallByParent(this.props,s))}}],[{key:"getDerivedStateFromProps",value:function(r,a){return r.animationId!==a.prevAnimationId?{prevAnimationId:r.animationId,curPoints:r.points,prevPoints:a.curPoints}:r.points!==a.curPoints?{curPoints:r.points}:null}},{key:"repeat",value:function(r,a){for(var o=r.length%2!==0?[].concat(Lo(r),[0]):r,i=[],s=0;s<a;++s)i=[].concat(Lo(i),Lo(o));return i}},{key:"renderDotItem",value:function(r,a){var o;if(E.isValidElement(r))o=E.cloneElement(r,a);else if(ae(r))o=r(a);else{var i=a.key,s=f2(a,are),l=ue("recharts-line-dot",typeof r!="boolean"?r.className:"");o=E.createElement(gd,cl({key:i},s,{className:l}))}return o}}])}(k.PureComponent);$n(Bt,"displayName","Line");$n(Bt,"defaultProps",{xAxisId:0,yAxisId:0,connectNulls:!1,activeDot:!0,dot:!0,legendType:"line",stroke:"#3182bd",strokeWidth:1,fill:"#fff",points:[],isAnimationActive:!_o.isSsr,animateNewValues:!0,animationBegin:0,animationDuration:1500,animationEasing:"ease",hide:!1,label:!1});$n(Bt,"getComposedData",function(e){var t=e.props,n=e.xAxis,r=e.yAxis,a=e.xAxisTicks,o=e.yAxisTicks,i=e.dataKey,s=e.bandSize,l=e.displayedData,u=e.offset,p=t.layout,c=l.map(function(f,m){var h=It(f,i);return p==="horizontal"?{x:Up({axis:n,ticks:a,bandSize:s,entry:f,index:m}),y:le(h)?null:r.scale(h),value:h,payload:f}:{x:le(h)?null:n.scale(h),y:Up({axis:r,ticks:o,bandSize:s,entry:f,index:m}),value:h,payload:f}});return Vt({points:c,layout:p},u)});var yre=["layout","type","stroke","connectNulls","isRange","ref"],gre=["key"],oS;function Yi(e){"@babel/helpers - typeof";return Yi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Yi(e)}function iS(e,t){if(e==null)return{};var n=xre(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function xre(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function Ja(){return Ja=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Ja.apply(this,arguments)}function h2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function Yr(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?h2(Object(n),!0).forEach(function(r){ar(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):h2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function wre(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function v2(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,lS(r.key),r)}}function bre(e,t,n){return t&&v2(e.prototype,t),n&&v2(e,n),Object.defineProperty(e,"prototype",{writable:!1}),e}function Pre(e,t,n){return t=ff(t),Sre(e,sS()?Reflect.construct(t,n||[],ff(e).constructor):t.apply(e,n))}function Sre(e,t){if(t&&(Yi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return Ore(e)}function Ore(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function sS(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(sS=function(){return!!e})()}function ff(e){return ff=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},ff(e)}function kre(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&cv(e,t)}function cv(e,t){return cv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},cv(e,t)}function ar(e,t,n){return t=lS(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function lS(e){var t=Cre(e,"string");return Yi(t)=="symbol"?t:t+""}function Cre(e,t){if(Yi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Yi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var ht=function(e){function t(){var n;wre(this,t);for(var r=arguments.length,a=new Array(r),o=0;o<r;o++)a[o]=arguments[o];return n=Pre(this,t,[].concat(a)),ar(n,"state",{isAnimationFinished:!0}),ar(n,"id",ys("recharts-area-")),ar(n,"handleAnimationEnd",function(){var i=n.props.onAnimationEnd;n.setState({isAnimationFinished:!0}),ae(i)&&i()}),ar(n,"handleAnimationStart",function(){var i=n.props.onAnimationStart;n.setState({isAnimationFinished:!1}),ae(i)&&i()}),n}return kre(t,e),bre(t,[{key:"renderDots",value:function(r,a,o){var i=this.props.isAnimationActive,s=this.state.isAnimationFinished;if(i&&!s)return null;var l=this.props,u=l.dot,p=l.points,c=l.dataKey,f=ie(this.props,!1),m=ie(u,!0),h=p.map(function(g,v){var y=Yr(Yr(Yr({key:"dot-".concat(v),r:3},f),m),{},{index:v,cx:g.x,cy:g.y,dataKey:c,value:g.value,payload:g.payload,points:p});return t.renderDotItem(u,y)}),d={clipPath:r?"url(#clipPath-".concat(a?"":"dots-").concat(o,")"):null};return E.createElement(Ne,Ja({className:"recharts-area-dots"},d),h)}},{key:"renderHorizontalRect",value:function(r){var a=this.props,o=a.baseLine,i=a.points,s=a.strokeWidth,l=i[0].x,u=i[i.length-1].x,p=r*Math.abs(l-u),c=pa(i.map(function(f){return f.y||0}));return V(o)&&typeof o=="number"?c=Math.max(o,c):o&&Array.isArray(o)&&o.length&&(c=Math.max(pa(o.map(function(f){return f.y||0})),c)),V(c)?E.createElement("rect",{x:l<u?l:l-p,y:0,width:p,height:Math.floor(c+(s?parseInt("".concat(s),10):1))}):null}},{key:"renderVerticalRect",value:function(r){var a=this.props,o=a.baseLine,i=a.points,s=a.strokeWidth,l=i[0].y,u=i[i.length-1].y,p=r*Math.abs(l-u),c=pa(i.map(function(f){return f.x||0}));return V(o)&&typeof o=="number"?c=Math.max(o,c):o&&Array.isArray(o)&&o.length&&(c=Math.max(pa(o.map(function(f){return f.x||0})),c)),V(c)?E.createElement("rect",{x:0,y:l<u?l:l-p,width:c+(s?parseInt("".concat(s),10):1),height:Math.floor(p)}):null}},{key:"renderClipRect",value:function(r){var a=this.props.layout;return a==="vertical"?this.renderVerticalRect(r):this.renderHorizontalRect(r)}},{key:"renderAreaStatically",value:function(r,a,o,i){var s=this.props,l=s.layout,u=s.type,p=s.stroke,c=s.connectNulls,f=s.isRange;s.ref;var m=iS(s,yre);return E.createElement(Ne,{clipPath:o?"url(#clipPath-".concat(i,")"):null},E.createElement(di,Ja({},ie(m,!0),{points:r,connectNulls:c,type:u,baseLine:a,layout:l,stroke:"none",className:"recharts-area-area"})),p!=="none"&&E.createElement(di,Ja({},ie(this.props,!1),{className:"recharts-area-curve",layout:l,type:u,connectNulls:c,fill:"none",points:r})),p!=="none"&&f&&E.createElement(di,Ja({},ie(this.props,!1),{className:"recharts-area-curve",layout:l,type:u,connectNulls:c,fill:"none",points:a})))}},{key:"renderAreaWithAnimation",value:function(r,a){var o=this,i=this.props,s=i.points,l=i.baseLine,u=i.isAnimationActive,p=i.animationBegin,c=i.animationDuration,f=i.animationEasing,m=i.animationId,h=this.state,d=h.prevPoints,g=h.prevBaseLine;return E.createElement(mr,{begin:p,duration:c,isActive:u,easing:f,from:{t:0},to:{t:1},key:"area-".concat(m),onAnimationEnd:this.handleAnimationEnd,onAnimationStart:this.handleAnimationStart},function(v){var y=v.t;if(d){var x=d.length/s.length,P=s.map(function(C,_){var T=Math.floor(_*x);if(d[T]){var A=d[T],j=dt(A.x,C.x),$=dt(A.y,C.y);return Yr(Yr({},C),{},{x:j(y),y:$(y)})}return C}),w;if(V(l)&&typeof l=="number"){var S=dt(g,l);w=S(y)}else if(le(l)||vs(l)){var O=dt(g,0);w=O(y)}else w=l.map(function(C,_){var T=Math.floor(_*x);if(g[T]){var A=g[T],j=dt(A.x,C.x),$=dt(A.y,C.y);return Yr(Yr({},C),{},{x:j(y),y:$(y)})}return C});return o.renderAreaStatically(P,w,r,a)}return E.createElement(Ne,null,E.createElement("defs",null,E.createElement("clipPath",{id:"animationClipPath-".concat(a)},o.renderClipRect(y))),E.createElement(Ne,{clipPath:"url(#animationClipPath-".concat(a,")")},o.renderAreaStatically(s,l,r,a)))})}},{key:"renderArea",value:function(r,a){var o=this.props,i=o.points,s=o.baseLine,l=o.isAnimationActive,u=this.state,p=u.prevPoints,c=u.prevBaseLine,f=u.totalLength;return l&&i&&i.length&&(!p&&f>0||!Fi(p,i)||!Fi(c,s))?this.renderAreaWithAnimation(r,a):this.renderAreaStatically(i,s,r,a)}},{key:"render",value:function(){var r,a=this.props,o=a.hide,i=a.dot,s=a.points,l=a.className,u=a.top,p=a.left,c=a.xAxis,f=a.yAxis,m=a.width,h=a.height,d=a.isAnimationActive,g=a.id;if(o||!s||!s.length)return null;var v=this.state.isAnimationFinished,y=s.length===1,x=ue("recharts-area",l),P=c&&c.allowDataOverflow,w=f&&f.allowDataOverflow,S=P||w,O=le(g)?this.id:g,C=(r=ie(i,!1))!==null&&r!==void 0?r:{r:3,strokeWidth:2},_=C.r,T=_===void 0?3:_,A=C.strokeWidth,j=A===void 0?2:A,$=a8(i)?i:{},M=$.clipDot,I=M===void 0?!0:M,R=T*2+j;return E.createElement(Ne,{className:x},P||w?E.createElement("defs",null,E.createElement("clipPath",{id:"clipPath-".concat(O)},E.createElement("rect",{x:P?p:p-m/2,y:w?u:u-h/2,width:P?m:m*2,height:w?h:h*2})),!I&&E.createElement("clipPath",{id:"clipPath-dots-".concat(O)},E.createElement("rect",{x:p-R/2,y:u-R/2,width:m+R,height:h+R}))):null,y?null:this.renderArea(S,O),(i||y)&&this.renderDots(S,I,O),(!d||v)&&Nr.renderCallByParent(this.props,s))}}],[{key:"getDerivedStateFromProps",value:function(r,a){return r.animationId!==a.prevAnimationId?{prevAnimationId:r.animationId,curPoints:r.points,curBaseLine:r.baseLine,prevPoints:a.curPoints,prevBaseLine:a.curBaseLine}:r.points!==a.curPoints||r.baseLine!==a.curBaseLine?{curPoints:r.points,curBaseLine:r.baseLine}:null}}])}(k.PureComponent);oS=ht;ar(ht,"displayName","Area");ar(ht,"defaultProps",{stroke:"#3182bd",fill:"#3182bd",fillOpacity:.6,xAxisId:0,yAxisId:0,legendType:"line",connectNulls:!1,points:[],dot:!1,activeDot:!0,hide:!1,isAnimationActive:!_o.isSsr,animationBegin:0,animationDuration:1500,animationEasing:"ease"});ar(ht,"getBaseValue",function(e,t,n,r){var a=e.layout,o=e.baseValue,i=t.props.baseValue,s=i??o;if(V(s)&&typeof s=="number")return s;var l=a==="horizontal"?r:n,u=l.scale.domain();if(l.type==="number"){var p=Math.max(u[0],u[1]),c=Math.min(u[0],u[1]);return s==="dataMin"?c:s==="dataMax"||p<0?p:Math.max(Math.min(u[0],u[1]),0)}return s==="dataMin"?u[0]:s==="dataMax"?u[1]:u[0]});ar(ht,"getComposedData",function(e){var t=e.props,n=e.item,r=e.xAxis,a=e.yAxis,o=e.xAxisTicks,i=e.yAxisTicks,s=e.bandSize,l=e.dataKey,u=e.stackedData,p=e.dataStartIndex,c=e.displayedData,f=e.offset,m=t.layout,h=u&&u.length,d=oS.getBaseValue(t,n,r,a),g=m==="horizontal",v=!1,y=c.map(function(P,w){var S;h?S=u[p+w]:(S=It(P,l),Array.isArray(S)?v=!0:S=[d,S]);var O=S[1]==null||h&&It(P,l)==null;return g?{x:Up({axis:r,ticks:o,bandSize:s,entry:P,index:w}),y:O?null:a.scale(S[1]),value:S,payload:P}:{x:O?null:r.scale(S[1]),y:Up({axis:a,ticks:i,bandSize:s,entry:P,index:w}),value:S,payload:P}}),x;return h||v?x=y.map(function(P){var w=Array.isArray(P.value)?P.value[0]:null;return g?{x:P.x,y:w!=null&&P.y!=null?a.scale(w):null}:{x:w!=null?r.scale(w):null,y:P.y}}):x=g?a.scale(d):r.scale(d),Yr({points:y,baseLine:x,layout:m,isRange:v},f)});ar(ht,"renderDotItem",function(e,t){var n;if(E.isValidElement(e))n=E.cloneElement(e,t);else if(ae(e))n=e(t);else{var r=ue("recharts-area-dot",typeof e!="boolean"?e.className:""),a=t.key,o=iS(t,gre);n=E.createElement(gd,Ja({},o,{key:a,className:r}))}return n});function Qi(e){"@babel/helpers - typeof";return Qi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Qi(e)}function _re(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Are(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,pS(r.key),r)}}function Ere(e,t,n){return t&&Are(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function Tre(e,t,n){return t=df(t),jre(e,uS()?Reflect.construct(t,n||[],df(e).constructor):t.apply(e,n))}function jre(e,t){if(t&&(Qi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return Nre(e)}function Nre(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function uS(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(uS=function(){return!!e})()}function df(e){return df=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},df(e)}function $re(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&pv(e,t)}function pv(e,t){return pv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},pv(e,t)}function cS(e,t,n){return t=pS(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function pS(e){var t=Mre(e,"string");return Qi(t)=="symbol"?t:t+""}function Mre(e,t){if(Qi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Qi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}function fv(){return fv=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},fv.apply(this,arguments)}function Rre(e){var t=e.xAxisId,n=$g(),r=Mg(),a=WP(t);return a==null?null:k.createElement(Os,fv({},a,{className:ue("recharts-".concat(a.axisType," ").concat(a.axisType),a.className),viewBox:{x:0,y:0,width:n,height:r},ticksGenerator:function(i){return _r(i,!0)}}))}var Mt=function(e){function t(){return _re(this,t),Tre(this,t,arguments)}return $re(t,e),Ere(t,[{key:"render",value:function(){return k.createElement(Rre,this.props)}}])}(k.Component);cS(Mt,"displayName","XAxis");cS(Mt,"defaultProps",{allowDecimals:!0,hide:!1,orientation:"bottom",width:0,height:30,mirror:!1,xAxisId:0,tickCount:5,type:"category",padding:{left:0,right:0},allowDataOverflow:!1,scale:"auto",reversed:!1,allowDuplicatedCategory:!0});function Zi(e){"@babel/helpers - typeof";return Zi=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Zi(e)}function Ire(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Dre(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,mS(r.key),r)}}function Lre(e,t,n){return t&&Dre(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function Fre(e,t,n){return t=mf(t),Bre(e,fS()?Reflect.construct(t,n||[],mf(e).constructor):t.apply(e,n))}function Bre(e,t){if(t&&(Zi(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return zre(e)}function zre(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function fS(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(fS=function(){return!!e})()}function mf(e){return mf=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},mf(e)}function Hre(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&dv(e,t)}function dv(e,t){return dv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},dv(e,t)}function dS(e,t,n){return t=mS(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function mS(e){var t=Gre(e,"string");return Zi(t)=="symbol"?t:t+""}function Gre(e,t){if(Zi(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Zi(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}function mv(){return mv=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},mv.apply(this,arguments)}var Ure=function(t){var n=t.yAxisId,r=$g(),a=Mg(),o=qP(n);return o==null?null:k.createElement(Os,mv({},o,{className:ue("recharts-".concat(o.axisType," ").concat(o.axisType),o.className),viewBox:{x:0,y:0,width:r,height:a},ticksGenerator:function(s){return _r(s,!0)}}))},_t=function(e){function t(){return Ire(this,t),Fre(this,t,arguments)}return Hre(t,e),Lre(t,[{key:"render",value:function(){return k.createElement(Ure,this.props)}}])}(k.Component);dS(_t,"displayName","YAxis");dS(_t,"defaultProps",{allowDuplicatedCategory:!0,allowDecimals:!0,hide:!1,orientation:"left",width:60,height:0,mirror:!1,yAxisId:0,tickCount:5,type:"number",padding:{top:0,bottom:0},allowDataOverflow:!1,scale:"auto",reversed:!1});function y2(e){return Kre(e)||Vre(e)||qre(e)||Wre()}function Wre(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function qre(e,t){if(e){if(typeof e=="string")return hv(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return hv(e,t)}}function Vre(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function Kre(e){if(Array.isArray(e))return hv(e)}function hv(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}var vv=function(t,n,r,a,o){var i=yn(t,Ig),s=yn(t,Sd),l=[].concat(y2(i),y2(s)),u=yn(t,kd),p="".concat(a,"Id"),c=a[0],f=n;if(l.length&&(f=l.reduce(function(d,g){if(g.props[p]===r&&pr(g.props,"extendDomain")&&V(g.props[c])){var v=g.props[c];return[Math.min(d[0],v),Math.max(d[1],v)]}return d},f)),u.length){var m="".concat(c,"1"),h="".concat(c,"2");f=u.reduce(function(d,g){if(g.props[p]===r&&pr(g.props,"extendDomain")&&V(g.props[m])&&V(g.props[h])){var v=g.props[m],y=g.props[h];return[Math.min(d[0],v,y),Math.max(d[1],v,y)]}return d},f)}return o&&o.length&&(f=o.reduce(function(d,g){return V(g)?[Math.min(d[0],g),Math.max(d[1],g)]:d},f)),f},hS={exports:{}};(function(e){var t=Object.prototype.hasOwnProperty,n="~";function r(){}Object.create&&(r.prototype=Object.create(null),new r().__proto__||(n=!1));function a(l,u,p){this.fn=l,this.context=u,this.once=p||!1}function o(l,u,p,c,f){if(typeof p!="function")throw new TypeError("The listener must be a function");var m=new a(p,c||l,f),h=n?n+u:u;return l._events[h]?l._events[h].fn?l._events[h]=[l._events[h],m]:l._events[h].push(m):(l._events[h]=m,l._eventsCount++),l}function i(l,u){--l._eventsCount===0?l._events=new r:delete l._events[u]}function s(){this._events=new r,this._eventsCount=0}s.prototype.eventNames=function(){var u=[],p,c;if(this._eventsCount===0)return u;for(c in p=this._events)t.call(p,c)&&u.push(n?c.slice(1):c);return Object.getOwnPropertySymbols?u.concat(Object.getOwnPropertySymbols(p)):u},s.prototype.listeners=function(u){var p=n?n+u:u,c=this._events[p];if(!c)return[];if(c.fn)return[c.fn];for(var f=0,m=c.length,h=new Array(m);f<m;f++)h[f]=c[f].fn;return h},s.prototype.listenerCount=function(u){var p=n?n+u:u,c=this._events[p];return c?c.fn?1:c.length:0},s.prototype.emit=function(u,p,c,f,m,h){var d=n?n+u:u;if(!this._events[d])return!1;var g=this._events[d],v=arguments.length,y,x;if(g.fn){switch(g.once&&this.removeListener(u,g.fn,void 0,!0),v){case 1:return g.fn.call(g.context),!0;case 2:return g.fn.call(g.context,p),!0;case 3:return g.fn.call(g.context,p,c),!0;case 4:return g.fn.call(g.context,p,c,f),!0;case 5:return g.fn.call(g.context,p,c,f,m),!0;case 6:return g.fn.call(g.context,p,c,f,m,h),!0}for(x=1,y=new Array(v-1);x<v;x++)y[x-1]=arguments[x];g.fn.apply(g.context,y)}else{var P=g.length,w;for(x=0;x<P;x++)switch(g[x].once&&this.removeListener(u,g[x].fn,void 0,!0),v){case 1:g[x].fn.call(g[x].context);break;case 2:g[x].fn.call(g[x].context,p);break;case 3:g[x].fn.call(g[x].context,p,c);break;case 4:g[x].fn.call(g[x].context,p,c,f);break;default:if(!y)for(w=1,y=new Array(v-1);w<v;w++)y[w-1]=arguments[w];g[x].fn.apply(g[x].context,y)}}return!0},s.prototype.on=function(u,p,c){return o(this,u,p,c,!1)},s.prototype.once=function(u,p,c){return o(this,u,p,c,!0)},s.prototype.removeListener=function(u,p,c,f){var m=n?n+u:u;if(!this._events[m])return this;if(!p)return i(this,m),this;var h=this._events[m];if(h.fn)h.fn===p&&(!f||h.once)&&(!c||h.context===c)&&i(this,m);else{for(var d=0,g=[],v=h.length;d<v;d++)(h[d].fn!==p||f&&!h[d].once||c&&h[d].context!==c)&&g.push(h[d]);g.length?this._events[m]=g.length===1?g[0]:g:i(this,m)}return this},s.prototype.removeAllListeners=function(u){var p;return u?(p=n?n+u:u,this._events[p]&&i(this,p)):(this._events=new r,this._eventsCount=0),this},s.prototype.off=s.prototype.removeListener,s.prototype.addListener=s.prototype.on,s.prefixed=n,s.EventEmitter=s,e.exports=s})(hS);var Xre=hS.exports;const Yre=_e(Xre);var zm=new Yre,Hm="recharts.syncMouseEvents";function cu(e){"@babel/helpers - typeof";return cu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},cu(e)}function Qre(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function Zre(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,vS(r.key),r)}}function Jre(e,t,n){return t&&Zre(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function Gm(e,t,n){return t=vS(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function vS(e){var t=eae(e,"string");return cu(t)=="symbol"?t:t+""}function eae(e,t){if(cu(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(cu(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return String(e)}var tae=function(){function e(){Qre(this,e),Gm(this,"activeIndex",0),Gm(this,"coordinateList",[]),Gm(this,"layout","horizontal")}return Jre(e,[{key:"setDetails",value:function(n){var r,a=n.coordinateList,o=a===void 0?null:a,i=n.container,s=i===void 0?null:i,l=n.layout,u=l===void 0?null:l,p=n.offset,c=p===void 0?null:p,f=n.mouseHandlerCallback,m=f===void 0?null:f;this.coordinateList=(r=o??this.coordinateList)!==null&&r!==void 0?r:[],this.container=s??this.container,this.layout=u??this.layout,this.offset=c??this.offset,this.mouseHandlerCallback=m??this.mouseHandlerCallback,this.activeIndex=Math.min(Math.max(this.activeIndex,0),this.coordinateList.length-1)}},{key:"focus",value:function(){this.spoofMouse()}},{key:"keyboardEvent",value:function(n){if(this.coordinateList.length!==0)switch(n.key){case"ArrowRight":{if(this.layout!=="horizontal")return;this.activeIndex=Math.min(this.activeIndex+1,this.coordinateList.length-1),this.spoofMouse();break}case"ArrowLeft":{if(this.layout!=="horizontal")return;this.activeIndex=Math.max(this.activeIndex-1,0),this.spoofMouse();break}}}},{key:"setIndex",value:function(n){this.activeIndex=n}},{key:"spoofMouse",value:function(){var n,r;if(this.layout==="horizontal"&&this.coordinateList.length!==0){var a=this.container.getBoundingClientRect(),o=a.x,i=a.y,s=a.height,l=this.coordinateList[this.activeIndex].coordinate,u=((n=window)===null||n===void 0?void 0:n.scrollX)||0,p=((r=window)===null||r===void 0?void 0:r.scrollY)||0,c=o+l+u,f=i+this.offset.top+s/2+p;this.mouseHandlerCallback({pageX:c,pageY:f})}}}])}();function nae(e,t,n){if(n==="number"&&t===!0&&Array.isArray(e)){var r=e==null?void 0:e[0],a=e==null?void 0:e[1];if(r&&a&&V(r)&&V(a))return!0}return!1}function rae(e,t,n,r){var a=r/2;return{stroke:"none",fill:"#ccc",x:e==="horizontal"?t.x-a:n.left+.5,y:e==="horizontal"?n.top+.5:t.y-a,width:e==="horizontal"?r:n.width-1,height:e==="horizontal"?n.height-1:r}}function yS(e){var t=e.cx,n=e.cy,r=e.radius,a=e.startAngle,o=e.endAngle,i=mt(t,n,r,a),s=mt(t,n,r,o);return{points:[i,s],cx:t,cy:n,radius:r,startAngle:a,endAngle:o}}function aae(e,t,n){var r,a,o,i;if(e==="horizontal")r=t.x,o=r,a=n.top,i=n.top+n.height;else if(e==="vertical")a=t.y,i=a,r=n.left,o=n.left+n.width;else if(t.cx!=null&&t.cy!=null)if(e==="centric"){var s=t.cx,l=t.cy,u=t.innerRadius,p=t.outerRadius,c=t.angle,f=mt(s,l,u,c),m=mt(s,l,p,c);r=f.x,a=f.y,o=m.x,i=m.y}else return yS(t);return[{x:r,y:a},{x:o,y:i}]}function pu(e){"@babel/helpers - typeof";return pu=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},pu(e)}function g2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function hc(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?g2(Object(n),!0).forEach(function(r){oae(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):g2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function oae(e,t,n){return t=iae(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function iae(e){var t=sae(e,"string");return pu(t)=="symbol"?t:t+""}function sae(e,t){if(pu(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(pu(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}function lae(e){var t,n,r=e.element,a=e.tooltipEventType,o=e.isActive,i=e.activeCoordinate,s=e.activePayload,l=e.offset,u=e.activeTooltipIndex,p=e.tooltipAxisBandSize,c=e.layout,f=e.chartName,m=(t=r.props.cursor)!==null&&t!==void 0?t:(n=r.type.defaultProps)===null||n===void 0?void 0:n.cursor;if(!r||!m||!o||!i||f!=="ScatterChart"&&a!=="axis")return null;var h,d=di;if(f==="ScatterChart")h=i,d=ZZ;else if(f==="BarChart")h=rae(c,i,l,p),d=Ag;else if(c==="radial"){var g=yS(i),v=g.cx,y=g.cy,x=g.radius,P=g.startAngle,w=g.endAngle;h={cx:v,cy:y,startAngle:P,endAngle:w,innerRadius:x,outerRadius:x},d=yP}else h={points:aae(c,i,l)},d=di;var S=hc(hc(hc(hc({stroke:"#ccc",pointerEvents:"none"},l),h),ie(m,!1)),{},{payload:s,payloadIndex:u,className:ue("recharts-tooltip-cursor",m.className)});return k.isValidElement(m)?k.cloneElement(m,S):k.createElement(d,S)}var uae=["item"],cae=["children","className","width","height","style","compact","title","desc"];function Ji(e){"@babel/helpers - typeof";return Ji=typeof Symbol=="function"&&typeof Symbol.iterator=="symbol"?function(t){return typeof t}:function(t){return t&&typeof Symbol=="function"&&t.constructor===Symbol&&t!==Symbol.prototype?"symbol":typeof t},Ji(e)}function Jo(){return Jo=Object.assign?Object.assign.bind():function(e){for(var t=1;t<arguments.length;t++){var n=arguments[t];for(var r in n)Object.prototype.hasOwnProperty.call(n,r)&&(e[r]=n[r])}return e},Jo.apply(this,arguments)}function x2(e,t){return dae(e)||fae(e,t)||xS(e,t)||pae()}function pae(){throw new TypeError(`Invalid attempt to destructure non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function fae(e,t){var n=e==null?null:typeof Symbol<"u"&&e[Symbol.iterator]||e["@@iterator"];if(n!=null){var r,a,o,i,s=[],l=!0,u=!1;try{if(o=(n=n.call(e)).next,t!==0)for(;!(l=(r=o.call(n)).done)&&(s.push(r.value),s.length!==t);l=!0);}catch(p){u=!0,a=p}finally{try{if(!l&&n.return!=null&&(i=n.return(),Object(i)!==i))return}finally{if(u)throw a}}return s}}function dae(e){if(Array.isArray(e))return e}function w2(e,t){if(e==null)return{};var n=mae(e,t),r,a;if(Object.getOwnPropertySymbols){var o=Object.getOwnPropertySymbols(e);for(a=0;a<o.length;a++)r=o[a],!(t.indexOf(r)>=0)&&Object.prototype.propertyIsEnumerable.call(e,r)&&(n[r]=e[r])}return n}function mae(e,t){if(e==null)return{};var n={};for(var r in e)if(Object.prototype.hasOwnProperty.call(e,r)){if(t.indexOf(r)>=0)continue;n[r]=e[r]}return n}function hae(e,t){if(!(e instanceof t))throw new TypeError("Cannot call a class as a function")}function vae(e,t){for(var n=0;n<t.length;n++){var r=t[n];r.enumerable=r.enumerable||!1,r.configurable=!0,"value"in r&&(r.writable=!0),Object.defineProperty(e,wS(r.key),r)}}function yae(e,t,n){return t&&vae(e.prototype,t),Object.defineProperty(e,"prototype",{writable:!1}),e}function gae(e,t,n){return t=hf(t),xae(e,gS()?Reflect.construct(t,n||[],hf(e).constructor):t.apply(e,n))}function xae(e,t){if(t&&(Ji(t)==="object"||typeof t=="function"))return t;if(t!==void 0)throw new TypeError("Derived constructors may only return object or undefined");return wae(e)}function wae(e){if(e===void 0)throw new ReferenceError("this hasn't been initialised - super() hasn't been called");return e}function gS(){try{var e=!Boolean.prototype.valueOf.call(Reflect.construct(Boolean,[],function(){}))}catch{}return(gS=function(){return!!e})()}function hf(e){return hf=Object.setPrototypeOf?Object.getPrototypeOf.bind():function(n){return n.__proto__||Object.getPrototypeOf(n)},hf(e)}function bae(e,t){if(typeof t!="function"&&t!==null)throw new TypeError("Super expression must either be null or a function");e.prototype=Object.create(t&&t.prototype,{constructor:{value:e,writable:!0,configurable:!0}}),Object.defineProperty(e,"prototype",{writable:!1}),t&&yv(e,t)}function yv(e,t){return yv=Object.setPrototypeOf?Object.setPrototypeOf.bind():function(r,a){return r.__proto__=a,r},yv(e,t)}function es(e){return Oae(e)||Sae(e)||xS(e)||Pae()}function Pae(){throw new TypeError(`Invalid attempt to spread non-iterable instance.
In order to be iterable, non-array objects must have a [Symbol.iterator]() method.`)}function xS(e,t){if(e){if(typeof e=="string")return gv(e,t);var n=Object.prototype.toString.call(e).slice(8,-1);if(n==="Object"&&e.constructor&&(n=e.constructor.name),n==="Map"||n==="Set")return Array.from(e);if(n==="Arguments"||/^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n))return gv(e,t)}}function Sae(e){if(typeof Symbol<"u"&&e[Symbol.iterator]!=null||e["@@iterator"]!=null)return Array.from(e)}function Oae(e){if(Array.isArray(e))return gv(e)}function gv(e,t){(t==null||t>e.length)&&(t=e.length);for(var n=0,r=new Array(t);n<t;n++)r[n]=e[n];return r}function b2(e,t){var n=Object.keys(e);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);t&&(r=r.filter(function(a){return Object.getOwnPropertyDescriptor(e,a).enumerable})),n.push.apply(n,r)}return n}function B(e){for(var t=1;t<arguments.length;t++){var n=arguments[t]!=null?arguments[t]:{};t%2?b2(Object(n),!0).forEach(function(r){ee(e,r,n[r])}):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(n)):b2(Object(n)).forEach(function(r){Object.defineProperty(e,r,Object.getOwnPropertyDescriptor(n,r))})}return e}function ee(e,t,n){return t=wS(t),t in e?Object.defineProperty(e,t,{value:n,enumerable:!0,configurable:!0,writable:!0}):e[t]=n,e}function wS(e){var t=kae(e,"string");return Ji(t)=="symbol"?t:t+""}function kae(e,t){if(Ji(e)!="object"||!e)return e;var n=e[Symbol.toPrimitive];if(n!==void 0){var r=n.call(e,t);if(Ji(r)!="object")return r;throw new TypeError("@@toPrimitive must return a primitive value.")}return(t==="string"?String:Number)(e)}var Cae={xAxis:["bottom","top"],yAxis:["left","right"]},_ae={width:"100%",height:"100%"},bS={x:0,y:0};function vc(e){return e}var Aae=function(t,n){return n==="horizontal"?t.x:n==="vertical"?t.y:n==="centric"?t.angle:t.radius},Eae=function(t,n,r,a){var o=n.find(function(p){return p&&p.index===r});if(o){if(t==="horizontal")return{x:o.coordinate,y:a.y};if(t==="vertical")return{x:a.x,y:o.coordinate};if(t==="centric"){var i=o.coordinate,s=a.radius;return B(B(B({},a),mt(a.cx,a.cy,s,i)),{},{angle:i,radius:s})}var l=o.coordinate,u=a.angle;return B(B(B({},a),mt(a.cx,a.cy,l,u)),{},{angle:u,radius:l})}return bS},Cd=function(t,n){var r=n.graphicalItems,a=n.dataStartIndex,o=n.dataEndIndex,i=(r??[]).reduce(function(s,l){var u=l.props.data;return u&&u.length?[].concat(es(s),es(u)):s},[]);return i.length>0?i:t&&t.length&&V(a)&&V(o)?t.slice(a,o+1):[]};function PS(e){return e==="number"?[0,"auto"]:void 0}var xv=function(t,n,r,a){var o=t.graphicalItems,i=t.tooltipAxis,s=Cd(n,t);return r<0||!o||!o.length||r>=s.length?null:o.reduce(function(l,u){var p,c=(p=u.props.data)!==null&&p!==void 0?p:n;c&&t.dataStartIndex+t.dataEndIndex!==0&&t.dataEndIndex-t.dataStartIndex>=r&&(c=c.slice(t.dataStartIndex,t.dataEndIndex+1));var f;if(i.dataKey&&!i.allowDuplicatedCategory){var m=c===void 0?s:c;f=mp(m,i.dataKey,a)}else f=c&&c[r]||s[r];return f?[].concat(es(l),[mP(u,f)]):l},[])},P2=function(t,n,r,a){var o=a||{x:t.chartX,y:t.chartY},i=Aae(o,r),s=t.orderedTooltipTicks,l=t.tooltipAxis,u=t.tooltipTicks,p=MX(i,s,u,l);if(p>=0&&u){var c=u[p]&&u[p].value,f=xv(t,n,p,c),m=Eae(r,s,p,o);return{activeTooltipIndex:p,activeLabel:c,activePayload:f,activeCoordinate:m}}return null},Tae=function(t,n){var r=n.axes,a=n.graphicalItems,o=n.axisType,i=n.axisIdKey,s=n.stackGroups,l=n.dataStartIndex,u=n.dataEndIndex,p=t.layout,c=t.children,f=t.stackOffset,m=pP(p,o);return r.reduce(function(h,d){var g,v=d.type.defaultProps!==void 0?B(B({},d.type.defaultProps),d.props):d.props,y=v.type,x=v.dataKey,P=v.allowDataOverflow,w=v.allowDuplicatedCategory,S=v.scale,O=v.ticks,C=v.includeHidden,_=v[i];if(h[_])return h;var T=Cd(t.data,{graphicalItems:a.filter(function(G){var Z,re=i in G.props?G.props[i]:(Z=G.type.defaultProps)===null||Z===void 0?void 0:Z[i];return re===_}),dataStartIndex:l,dataEndIndex:u}),A=T.length,j,$,M;nae(v.domain,P,y)&&(j=I0(v.domain,null,P),m&&(y==="number"||S!=="auto")&&(M=ll(T,x,"category")));var I=PS(y);if(!j||j.length===0){var R,L=(R=v.domain)!==null&&R!==void 0?R:I;if(x){if(j=ll(T,x,y),y==="category"&&m){var N=bM(j);w&&N?($=j,j=ef(0,A)):w||(j=Z3(L,j,d).reduce(function(G,Z){return G.indexOf(Z)>=0?G:[].concat(es(G),[Z])},[]))}else if(y==="category")w?j=j.filter(function(G){return G!==""&&!le(G)}):j=Z3(L,j,d).reduce(function(G,Z){return G.indexOf(Z)>=0||Z===""||le(Z)?G:[].concat(es(G),[Z])},[]);else if(y==="number"){var D=FX(T,a.filter(function(G){var Z,re,ve=i in G.props?G.props[i]:(Z=G.type.defaultProps)===null||Z===void 0?void 0:Z[i],be="hide"in G.props?G.props.hide:(re=G.type.defaultProps)===null||re===void 0?void 0:re.hide;return ve===_&&(C||!be)}),x,o,p);D&&(j=D)}m&&(y==="number"||S!=="auto")&&(M=ll(T,x,"category"))}else m?j=ef(0,A):s&&s[_]&&s[_].hasStack&&y==="number"?j=f==="expand"?[0,1]:dP(s[_].stackGroups,l,u):j=cP(T,a.filter(function(G){var Z=i in G.props?G.props[i]:G.type.defaultProps[i],re="hide"in G.props?G.props.hide:G.type.defaultProps.hide;return Z===_&&(C||!re)}),y,p,!0);if(y==="number")j=vv(c,j,_,o,O),L&&(j=I0(L,j,P));else if(y==="category"&&L){var H=L,W=j.every(function(G){return H.indexOf(G)>=0});W&&(j=H)}}return B(B({},h),{},ee({},_,B(B({},v),{},{axisType:o,domain:j,categoricalDomain:M,duplicateDomain:$,originalDomain:(g=v.domain)!==null&&g!==void 0?g:I,isCategorical:m,layout:p})))},{})},jae=function(t,n){var r=n.graphicalItems,a=n.Axis,o=n.axisType,i=n.axisIdKey,s=n.stackGroups,l=n.dataStartIndex,u=n.dataEndIndex,p=t.layout,c=t.children,f=Cd(t.data,{graphicalItems:r,dataStartIndex:l,dataEndIndex:u}),m=f.length,h=pP(p,o),d=-1;return r.reduce(function(g,v){var y=v.type.defaultProps!==void 0?B(B({},v.type.defaultProps),v.props):v.props,x=y[i],P=PS("number");if(!g[x]){d++;var w;return h?w=ef(0,m):s&&s[x]&&s[x].hasStack?(w=dP(s[x].stackGroups,l,u),w=vv(c,w,x,o)):(w=I0(P,cP(f,r.filter(function(S){var O,C,_=i in S.props?S.props[i]:(O=S.type.defaultProps)===null||O===void 0?void 0:O[i],T="hide"in S.props?S.props.hide:(C=S.type.defaultProps)===null||C===void 0?void 0:C.hide;return _===x&&!T}),"number",p),a.defaultProps.allowDataOverflow),w=vv(c,w,x,o)),B(B({},g),{},ee({},x,B(B({axisType:o},a.defaultProps),{},{hide:!0,orientation:vn(Cae,"".concat(o,".").concat(d%2),null),domain:w,originalDomain:P,isCategorical:h,layout:p})))}return g},{})},Nae=function(t,n){var r=n.axisType,a=r===void 0?"xAxis":r,o=n.AxisComp,i=n.graphicalItems,s=n.stackGroups,l=n.dataStartIndex,u=n.dataEndIndex,p=t.children,c="".concat(a,"Id"),f=yn(p,o),m={};return f&&f.length?m=Tae(t,{axes:f,graphicalItems:i,axisType:a,axisIdKey:c,stackGroups:s,dataStartIndex:l,dataEndIndex:u}):i&&i.length&&(m=jae(t,{Axis:o,graphicalItems:i,axisType:a,axisIdKey:c,stackGroups:s,dataStartIndex:l,dataEndIndex:u})),m},$ae=function(t){var n=ra(t),r=_r(n,!1,!0);return{tooltipTicks:r,orderedTooltipTicks:tg(r,function(a){return a.coordinate}),tooltipAxis:n,tooltipAxisBandSize:Wp(n,r)}},S2=function(t){var n=t.children,r=t.defaultShowTooltip,a=Yt(n,Gi),o=0,i=0;return t.data&&t.data.length!==0&&(i=t.data.length-1),a&&a.props&&(a.props.startIndex>=0&&(o=a.props.startIndex),a.props.endIndex>=0&&(i=a.props.endIndex)),{chartX:0,chartY:0,dataStartIndex:o,dataEndIndex:i,activeTooltipIndex:-1,isTooltipActive:!!r}},Mae=function(t){return!t||!t.length?!1:t.some(function(n){var r=Er(n&&n.type);return r&&r.indexOf("Bar")>=0})},O2=function(t){return t==="horizontal"?{numericAxisName:"yAxis",cateAxisName:"xAxis"}:t==="vertical"?{numericAxisName:"xAxis",cateAxisName:"yAxis"}:t==="centric"?{numericAxisName:"radiusAxis",cateAxisName:"angleAxis"}:{numericAxisName:"angleAxis",cateAxisName:"radiusAxis"}},Rae=function(t,n){var r=t.props,a=t.graphicalItems,o=t.xAxisMap,i=o===void 0?{}:o,s=t.yAxisMap,l=s===void 0?{}:s,u=r.width,p=r.height,c=r.children,f=r.margin||{},m=Yt(c,Gi),h=Yt(c,ci),d=Object.keys(l).reduce(function(w,S){var O=l[S],C=O.orientation;return!O.mirror&&!O.hide?B(B({},w),{},ee({},C,w[C]+O.width)):w},{left:f.left||0,right:f.right||0}),g=Object.keys(i).reduce(function(w,S){var O=i[S],C=O.orientation;return!O.mirror&&!O.hide?B(B({},w),{},ee({},C,vn(w,"".concat(C))+O.height)):w},{top:f.top||0,bottom:f.bottom||0}),v=B(B({},g),d),y=v.bottom;m&&(v.bottom+=m.props.height||Gi.defaultProps.height),h&&n&&(v=DX(v,a,r,n));var x=u-v.left-v.right,P=p-v.top-v.bottom;return B(B({brushBottom:y},v),{},{width:Math.max(x,0),height:Math.max(P,0)})},Iae=function(t,n){if(n==="xAxis")return t[n].width;if(n==="yAxis")return t[n].height},Fg=function(t){var n=t.chartName,r=t.GraphicalChild,a=t.defaultTooltipEventType,o=a===void 0?"axis":a,i=t.validateTooltipEventTypes,s=i===void 0?["axis"]:i,l=t.axisComponents,u=t.legendContent,p=t.formatAxisMap,c=t.defaultProps,f=function(v,y){var x=y.graphicalItems,P=y.stackGroups,w=y.offset,S=y.updateId,O=y.dataStartIndex,C=y.dataEndIndex,_=v.barSize,T=v.layout,A=v.barGap,j=v.barCategoryGap,$=v.maxBarSize,M=O2(T),I=M.numericAxisName,R=M.cateAxisName,L=Mae(x),N=[];return x.forEach(function(D,H){var W=Cd(v.data,{graphicalItems:[D],dataStartIndex:O,dataEndIndex:C}),G=D.type.defaultProps!==void 0?B(B({},D.type.defaultProps),D.props):D.props,Z=G.dataKey,re=G.maxBarSize,ve=G["".concat(I,"Id")],be=G["".concat(R,"Id")],J={},se=l.reduce(function(La,Fa){var Nd=y["".concat(Fa.axisType,"Map")],Kg=G["".concat(Fa.axisType,"Id")];Nd&&Nd[Kg]||Fa.axisType==="zAxis"||Po();var Xg=Nd[Kg];return B(B({},La),{},ee(ee({},Fa.axisType,Xg),"".concat(Fa.axisType,"Ticks"),_r(Xg)))},J),q=se[R],K=se["".concat(R,"Ticks")],X=P&&P[ve]&&P[ve].hasStack&&QX(D,P[ve].stackGroups),F=Er(D.type).indexOf("Bar")>=0,pe=Wp(q,K),te=[],$e=L&&RX({barSize:_,stackGroups:P,totalSize:Iae(se,R)});if(F){var Me,Qe,Vn=le(re)?$:re,Sn=(Me=(Qe=Wp(q,K,!0))!==null&&Qe!==void 0?Qe:Vn)!==null&&Me!==void 0?Me:0;te=IX({barGap:A,barCategoryGap:j,bandSize:Sn!==pe?Sn:pe,sizeList:$e[be],maxBarSize:Vn}),Sn!==pe&&(te=te.map(function(La){return B(B({},La),{},{position:B(B({},La.position),{},{offset:La.position.offset-Sn/2})})}))}var Nu=D&&D.type&&D.type.getComposedData;Nu&&N.push({props:B(B({},Nu(B(B({},se),{},{displayedData:W,props:v,dataKey:Z,item:D,bandSize:pe,barPosition:te,offset:w,stackedData:X,layout:T,dataStartIndex:O,dataEndIndex:C}))),{},ee(ee(ee({key:D.key||"item-".concat(H)},I,se[I]),R,se[R]),"animationId",S)),childIndex:$M(D,v.children),item:D})}),N},m=function(v,y){var x=v.props,P=v.dataStartIndex,w=v.dataEndIndex,S=v.updateId;if(!qx({props:x}))return null;var O=x.children,C=x.layout,_=x.stackOffset,T=x.data,A=x.reverseStackOrder,j=O2(C),$=j.numericAxisName,M=j.cateAxisName,I=yn(O,r),R=KX(T,I,"".concat($,"Id"),"".concat(M,"Id"),_,A),L=l.reduce(function(G,Z){var re="".concat(Z.axisType,"Map");return B(B({},G),{},ee({},re,Nae(x,B(B({},Z),{},{graphicalItems:I,stackGroups:Z.axisType===$&&R,dataStartIndex:P,dataEndIndex:w}))))},{}),N=Rae(B(B({},L),{},{props:x,graphicalItems:I}),y==null?void 0:y.legendBBox);Object.keys(L).forEach(function(G){L[G]=p(x,L[G],N,G.replace("Map",""),n)});var D=L["".concat(M,"Map")],H=$ae(D),W=f(x,B(B({},L),{},{dataStartIndex:P,dataEndIndex:w,updateId:S,graphicalItems:I,stackGroups:R,offset:N}));return B(B({formattedGraphicalItems:W,graphicalItems:I,offset:N,stackGroups:R},H),L)},h=function(g){function v(y){var x,P,w;return hae(this,v),w=gae(this,v,[y]),ee(w,"eventEmitterSymbol",Symbol("rechartsEventEmitter")),ee(w,"accessibilityManager",new tae),ee(w,"handleLegendBBoxUpdate",function(S){if(S){var O=w.state,C=O.dataStartIndex,_=O.dataEndIndex,T=O.updateId;w.setState(B({legendBBox:S},m({props:w.props,dataStartIndex:C,dataEndIndex:_,updateId:T},B(B({},w.state),{},{legendBBox:S}))))}}),ee(w,"handleReceiveSyncEvent",function(S,O,C){if(w.props.syncId===S){if(C===w.eventEmitterSymbol&&typeof w.props.syncMethod!="function")return;w.applySyncEvent(O)}}),ee(w,"handleBrushChange",function(S){var O=S.startIndex,C=S.endIndex;if(O!==w.state.dataStartIndex||C!==w.state.dataEndIndex){var _=w.state.updateId;w.setState(function(){return B({dataStartIndex:O,dataEndIndex:C},m({props:w.props,dataStartIndex:O,dataEndIndex:C,updateId:_},w.state))}),w.triggerSyncEvent({dataStartIndex:O,dataEndIndex:C})}}),ee(w,"handleMouseEnter",function(S){var O=w.getMouseInfo(S);if(O){var C=B(B({},O),{},{isTooltipActive:!0});w.setState(C),w.triggerSyncEvent(C);var _=w.props.onMouseEnter;ae(_)&&_(C,S)}}),ee(w,"triggeredAfterMouseMove",function(S){var O=w.getMouseInfo(S),C=O?B(B({},O),{},{isTooltipActive:!0}):{isTooltipActive:!1};w.setState(C),w.triggerSyncEvent(C);var _=w.props.onMouseMove;ae(_)&&_(C,S)}),ee(w,"handleItemMouseEnter",function(S){w.setState(function(){return{isTooltipActive:!0,activeItem:S,activePayload:S.tooltipPayload,activeCoordinate:S.tooltipPosition||{x:S.cx,y:S.cy}}})}),ee(w,"handleItemMouseLeave",function(){w.setState(function(){return{isTooltipActive:!1}})}),ee(w,"handleMouseMove",function(S){S.persist(),w.throttleTriggeredAfterMouseMove(S)}),ee(w,"handleMouseLeave",function(S){w.throttleTriggeredAfterMouseMove.cancel();var O={isTooltipActive:!1};w.setState(O),w.triggerSyncEvent(O);var C=w.props.onMouseLeave;ae(C)&&C(O,S)}),ee(w,"handleOuterEvent",function(S){var O=NM(S),C=vn(w.props,"".concat(O));if(O&&ae(C)){var _,T;/.*touch.*/i.test(O)?T=w.getMouseInfo(S.changedTouches[0]):T=w.getMouseInfo(S),C((_=T)!==null&&_!==void 0?_:{},S)}}),ee(w,"handleClick",function(S){var O=w.getMouseInfo(S);if(O){var C=B(B({},O),{},{isTooltipActive:!0});w.setState(C),w.triggerSyncEvent(C);var _=w.props.onClick;ae(_)&&_(C,S)}}),ee(w,"handleMouseDown",function(S){var O=w.props.onMouseDown;if(ae(O)){var C=w.getMouseInfo(S);O(C,S)}}),ee(w,"handleMouseUp",function(S){var O=w.props.onMouseUp;if(ae(O)){var C=w.getMouseInfo(S);O(C,S)}}),ee(w,"handleTouchMove",function(S){S.changedTouches!=null&&S.changedTouches.length>0&&w.throttleTriggeredAfterMouseMove(S.changedTouches[0])}),ee(w,"handleTouchStart",function(S){S.changedTouches!=null&&S.changedTouches.length>0&&w.handleMouseDown(S.changedTouches[0])}),ee(w,"handleTouchEnd",function(S){S.changedTouches!=null&&S.changedTouches.length>0&&w.handleMouseUp(S.changedTouches[0])}),ee(w,"handleDoubleClick",function(S){var O=w.props.onDoubleClick;if(ae(O)){var C=w.getMouseInfo(S);O(C,S)}}),ee(w,"handleContextMenu",function(S){var O=w.props.onContextMenu;if(ae(O)){var C=w.getMouseInfo(S);O(C,S)}}),ee(w,"triggerSyncEvent",function(S){w.props.syncId!==void 0&&zm.emit(Hm,w.props.syncId,S,w.eventEmitterSymbol)}),ee(w,"applySyncEvent",function(S){var O=w.props,C=O.layout,_=O.syncMethod,T=w.state.updateId,A=S.dataStartIndex,j=S.dataEndIndex;if(S.dataStartIndex!==void 0||S.dataEndIndex!==void 0)w.setState(B({dataStartIndex:A,dataEndIndex:j},m({props:w.props,dataStartIndex:A,dataEndIndex:j,updateId:T},w.state)));else if(S.activeTooltipIndex!==void 0){var $=S.chartX,M=S.chartY,I=S.activeTooltipIndex,R=w.state,L=R.offset,N=R.tooltipTicks;if(!L)return;if(typeof _=="function")I=_(N,S);else if(_==="value"){I=-1;for(var D=0;D<N.length;D++)if(N[D].value===S.activeLabel){I=D;break}}var H=B(B({},L),{},{x:L.left,y:L.top}),W=Math.min($,H.x+H.width),G=Math.min(M,H.y+H.height),Z=N[I]&&N[I].value,re=xv(w.state,w.props.data,I),ve=N[I]?{x:C==="horizontal"?N[I].coordinate:W,y:C==="horizontal"?G:N[I].coordinate}:bS;w.setState(B(B({},S),{},{activeLabel:Z,activeCoordinate:ve,activePayload:re,activeTooltipIndex:I}))}else w.setState(S)}),ee(w,"renderCursor",function(S){var O,C=w.state,_=C.isTooltipActive,T=C.activeCoordinate,A=C.activePayload,j=C.offset,$=C.activeTooltipIndex,M=C.tooltipAxisBandSize,I=w.getTooltipEventType(),R=(O=S.props.active)!==null&&O!==void 0?O:_,L=w.props.layout,N=S.key||"_recharts-cursor";return E.createElement(lae,{key:N,activeCoordinate:T,activePayload:A,activeTooltipIndex:$,chartName:n,element:S,isActive:R,layout:L,offset:j,tooltipAxisBandSize:M,tooltipEventType:I})}),ee(w,"renderPolarAxis",function(S,O,C){var _=vn(S,"type.axisType"),T=vn(w.state,"".concat(_,"Map")),A=S.type.defaultProps,j=A!==void 0?B(B({},A),S.props):S.props,$=T&&T[j["".concat(_,"Id")]];return k.cloneElement(S,B(B({},$),{},{className:ue(_,$.className),key:S.key||"".concat(O,"-").concat(C),ticks:_r($,!0)}))}),ee(w,"renderPolarGrid",function(S){var O=S.props,C=O.radialLines,_=O.polarAngles,T=O.polarRadius,A=w.state,j=A.radiusAxisMap,$=A.angleAxisMap,M=ra(j),I=ra($),R=I.cx,L=I.cy,N=I.innerRadius,D=I.outerRadius;return k.cloneElement(S,{polarAngles:Array.isArray(_)?_:_r(I,!0).map(function(H){return H.coordinate}),polarRadius:Array.isArray(T)?T:_r(M,!0).map(function(H){return H.coordinate}),cx:R,cy:L,innerRadius:N,outerRadius:D,key:S.key||"polar-grid",radialLines:C})}),ee(w,"renderLegend",function(){var S=w.state.formattedGraphicalItems,O=w.props,C=O.children,_=O.width,T=O.height,A=w.props.margin||{},j=_-(A.left||0)-(A.right||0),$=lP({children:C,formattedGraphicalItems:S,legendWidth:j,legendContent:u});if(!$)return null;var M=$.item,I=w2($,uae);return k.cloneElement(M,B(B({},I),{},{chartWidth:_,chartHeight:T,margin:A,onBBoxUpdate:w.handleLegendBBoxUpdate}))}),ee(w,"renderTooltip",function(){var S,O=w.props,C=O.children,_=O.accessibilityLayer,T=Yt(C,Yn);if(!T)return null;var A=w.state,j=A.isTooltipActive,$=A.activeCoordinate,M=A.activePayload,I=A.activeLabel,R=A.offset,L=(S=T.props.active)!==null&&S!==void 0?S:j;return k.cloneElement(T,{viewBox:B(B({},R),{},{x:R.left,y:R.top}),active:L,label:I,payload:L?M:[],coordinate:$,accessibilityLayer:_})}),ee(w,"renderBrush",function(S){var O=w.props,C=O.margin,_=O.data,T=w.state,A=T.offset,j=T.dataStartIndex,$=T.dataEndIndex,M=T.updateId;return k.cloneElement(S,{key:S.key||"_recharts-brush",onChange:pc(w.handleBrushChange,S.props.onChange),data:_,x:V(S.props.x)?S.props.x:A.left,y:V(S.props.y)?S.props.y:A.top+A.height+A.brushBottom-(C.bottom||0),width:V(S.props.width)?S.props.width:A.width,startIndex:j,endIndex:$,updateId:"brush-".concat(M)})}),ee(w,"renderReferenceElement",function(S,O,C){if(!S)return null;var _=w,T=_.clipPathId,A=w.state,j=A.xAxisMap,$=A.yAxisMap,M=A.offset,I=S.type.defaultProps||{},R=S.props,L=R.xAxisId,N=L===void 0?I.xAxisId:L,D=R.yAxisId,H=D===void 0?I.yAxisId:D;return k.cloneElement(S,{key:S.key||"".concat(O,"-").concat(C),xAxis:j[N],yAxis:$[H],viewBox:{x:M.left,y:M.top,width:M.width,height:M.height},clipPathId:T})}),ee(w,"renderActivePoints",function(S){var O=S.item,C=S.activePoint,_=S.basePoint,T=S.childIndex,A=S.isRange,j=[],$=O.props.key,M=O.item.type.defaultProps!==void 0?B(B({},O.item.type.defaultProps),O.item.props):O.item.props,I=M.activeDot,R=M.dataKey,L=B(B({index:T,dataKey:R,cx:C.x,cy:C.y,r:4,fill:_g(O.item),strokeWidth:2,stroke:"#fff",payload:C.payload,value:C.value},ie(I,!1)),hp(I));return j.push(v.renderActiveDot(I,L,"".concat($,"-activePoint-").concat(T))),_?j.push(v.renderActiveDot(I,B(B({},L),{},{cx:_.x,cy:_.y}),"".concat($,"-basePoint-").concat(T))):A&&j.push(null),j}),ee(w,"renderGraphicChild",function(S,O,C){var _=w.filterFormatItem(S,O,C);if(!_)return null;var T=w.getTooltipEventType(),A=w.state,j=A.isTooltipActive,$=A.tooltipAxis,M=A.activeTooltipIndex,I=A.activeLabel,R=w.props.children,L=Yt(R,Yn),N=_.props,D=N.points,H=N.isRange,W=N.baseLine,G=_.item.type.defaultProps!==void 0?B(B({},_.item.type.defaultProps),_.item.props):_.item.props,Z=G.activeDot,re=G.hide,ve=G.activeBar,be=G.activeShape,J=!!(!re&&j&&L&&(Z||ve||be)),se={};T!=="axis"&&L&&L.props.trigger==="click"?se={onClick:pc(w.handleItemMouseEnter,S.props.onClick)}:T!=="axis"&&(se={onMouseLeave:pc(w.handleItemMouseLeave,S.props.onMouseLeave),onMouseEnter:pc(w.handleItemMouseEnter,S.props.onMouseEnter)});var q=k.cloneElement(S,B(B({},_.props),se));function K(Fa){return typeof $.dataKey=="function"?$.dataKey(Fa.payload):null}if(J)if(M>=0){var X,F;if($.dataKey&&!$.allowDuplicatedCategory){var pe=typeof $.dataKey=="function"?K:"payload.".concat($.dataKey.toString());X=mp(D,pe,I),F=H&&W&&mp(W,pe,I)}else X=D==null?void 0:D[M],F=H&&W&&W[M];if(be||ve){var te=S.props.activeIndex!==void 0?S.props.activeIndex:M;return[k.cloneElement(S,B(B(B({},_.props),se),{},{activeIndex:te})),null,null]}if(!le(X))return[q].concat(es(w.renderActivePoints({item:_,activePoint:X,basePoint:F,childIndex:M,isRange:H})))}else{var $e,Me=($e=w.getItemByXY(w.state.activeCoordinate))!==null&&$e!==void 0?$e:{graphicalItem:q},Qe=Me.graphicalItem,Vn=Qe.item,Sn=Vn===void 0?S:Vn,Nu=Qe.childIndex,La=B(B(B({},_.props),se),{},{activeIndex:Nu});return[k.cloneElement(Sn,La),null,null]}return H?[q,null,null]:[q,null]}),ee(w,"renderCustomized",function(S,O,C){return k.cloneElement(S,B(B({key:"recharts-customized-".concat(C)},w.props),w.state))}),ee(w,"renderMap",{CartesianGrid:{handler:vc,once:!0},ReferenceArea:{handler:w.renderReferenceElement},ReferenceLine:{handler:vc},ReferenceDot:{handler:w.renderReferenceElement},XAxis:{handler:vc},YAxis:{handler:vc},Brush:{handler:w.renderBrush,once:!0},Bar:{handler:w.renderGraphicChild},Line:{handler:w.renderGraphicChild},Area:{handler:w.renderGraphicChild},Radar:{handler:w.renderGraphicChild},RadialBar:{handler:w.renderGraphicChild},Scatter:{handler:w.renderGraphicChild},Pie:{handler:w.renderGraphicChild},Funnel:{handler:w.renderGraphicChild},Tooltip:{handler:w.renderCursor,once:!0},PolarGrid:{handler:w.renderPolarGrid,once:!0},PolarAngleAxis:{handler:w.renderPolarAxis},PolarRadiusAxis:{handler:w.renderPolarAxis},Customized:{handler:w.renderCustomized}}),w.clipPathId="".concat((x=y.id)!==null&&x!==void 0?x:ys("recharts"),"-clip"),w.throttleTriggeredAfterMouseMove=s7(w.triggeredAfterMouseMove,(P=y.throttleDelay)!==null&&P!==void 0?P:1e3/60),w.state={},w}return bae(v,g),yae(v,[{key:"componentDidMount",value:function(){var x,P;this.addListener(),this.accessibilityManager.setDetails({container:this.container,offset:{left:(x=this.props.margin.left)!==null&&x!==void 0?x:0,top:(P=this.props.margin.top)!==null&&P!==void 0?P:0},coordinateList:this.state.tooltipTicks,mouseHandlerCallback:this.triggeredAfterMouseMove,layout:this.props.layout}),this.displayDefaultTooltip()}},{key:"displayDefaultTooltip",value:function(){var x=this.props,P=x.children,w=x.data,S=x.height,O=x.layout,C=Yt(P,Yn);if(C){var _=C.props.defaultIndex;if(!(typeof _!="number"||_<0||_>this.state.tooltipTicks.length-1)){var T=this.state.tooltipTicks[_]&&this.state.tooltipTicks[_].value,A=xv(this.state,w,_,T),j=this.state.tooltipTicks[_].coordinate,$=(this.state.offset.top+S)/2,M=O==="horizontal",I=M?{x:j,y:$}:{y:j,x:$},R=this.state.formattedGraphicalItems.find(function(N){var D=N.item;return D.type.name==="Scatter"});R&&(I=B(B({},I),R.props.points[_].tooltipPosition),A=R.props.points[_].tooltipPayload);var L={activeTooltipIndex:_,isTooltipActive:!0,activeLabel:T,activePayload:A,activeCoordinate:I};this.setState(L),this.renderCursor(C),this.accessibilityManager.setIndex(_)}}}},{key:"getSnapshotBeforeUpdate",value:function(x,P){if(!this.props.accessibilityLayer)return null;if(this.state.tooltipTicks!==P.tooltipTicks&&this.accessibilityManager.setDetails({coordinateList:this.state.tooltipTicks}),this.props.layout!==x.layout&&this.accessibilityManager.setDetails({layout:this.props.layout}),this.props.margin!==x.margin){var w,S;this.accessibilityManager.setDetails({offset:{left:(w=this.props.margin.left)!==null&&w!==void 0?w:0,top:(S=this.props.margin.top)!==null&&S!==void 0?S:0}})}return null}},{key:"componentDidUpdate",value:function(x){Qh([Yt(x.children,Yn)],[Yt(this.props.children,Yn)])||this.displayDefaultTooltip()}},{key:"componentWillUnmount",value:function(){this.removeListener(),this.throttleTriggeredAfterMouseMove.cancel()}},{key:"getTooltipEventType",value:function(){var x=Yt(this.props.children,Yn);if(x&&typeof x.props.shared=="boolean"){var P=x.props.shared?"axis":"item";return s.indexOf(P)>=0?P:o}return o}},{key:"getMouseInfo",value:function(x){if(!this.container)return null;var P=this.container,w=P.getBoundingClientRect(),S=iW(w),O={chartX:Math.round(x.pageX-S.left),chartY:Math.round(x.pageY-S.top)},C=w.width/P.offsetWidth||1,_=this.inRange(O.chartX,O.chartY,C);if(!_)return null;var T=this.state,A=T.xAxisMap,j=T.yAxisMap,$=this.getTooltipEventType(),M=P2(this.state,this.props.data,this.props.layout,_);if($!=="axis"&&A&&j){var I=ra(A).scale,R=ra(j).scale,L=I&&I.invert?I.invert(O.chartX):null,N=R&&R.invert?R.invert(O.chartY):null;return B(B({},O),{},{xValue:L,yValue:N},M)}return M?B(B({},O),M):null}},{key:"inRange",value:function(x,P){var w=arguments.length>2&&arguments[2]!==void 0?arguments[2]:1,S=this.props.layout,O=x/w,C=P/w;if(S==="horizontal"||S==="vertical"){var _=this.state.offset,T=O>=_.left&&O<=_.left+_.width&&C>=_.top&&C<=_.top+_.height;return T?{x:O,y:C}:null}var A=this.state,j=A.angleAxisMap,$=A.radiusAxisMap;if(j&&$){var M=ra(j);return tb({x:O,y:C},M)}return null}},{key:"parseEventsOfWrapper",value:function(){var x=this.props.children,P=this.getTooltipEventType(),w=Yt(x,Yn),S={};w&&P==="axis"&&(w.props.trigger==="click"?S={onClick:this.handleClick}:S={onMouseEnter:this.handleMouseEnter,onDoubleClick:this.handleDoubleClick,onMouseMove:this.handleMouseMove,onMouseLeave:this.handleMouseLeave,onTouchMove:this.handleTouchMove,onTouchStart:this.handleTouchStart,onTouchEnd:this.handleTouchEnd,onContextMenu:this.handleContextMenu});var O=hp(this.props,this.handleOuterEvent);return B(B({},O),S)}},{key:"addListener",value:function(){zm.on(Hm,this.handleReceiveSyncEvent)}},{key:"removeListener",value:function(){zm.removeListener(Hm,this.handleReceiveSyncEvent)}},{key:"filterFormatItem",value:function(x,P,w){for(var S=this.state.formattedGraphicalItems,O=0,C=S.length;O<C;O++){var _=S[O];if(_.item===x||_.props.key===x.key||P===Er(_.item.type)&&w===_.childIndex)return _}return null}},{key:"renderClipPath",value:function(){var x=this.clipPathId,P=this.state.offset,w=P.left,S=P.top,O=P.height,C=P.width;return E.createElement("defs",null,E.createElement("clipPath",{id:x},E.createElement("rect",{x:w,y:S,height:O,width:C})))}},{key:"getXScales",value:function(){var x=this.state.xAxisMap;return x?Object.entries(x).reduce(function(P,w){var S=x2(w,2),O=S[0],C=S[1];return B(B({},P),{},ee({},O,C.scale))},{}):null}},{key:"getYScales",value:function(){var x=this.state.yAxisMap;return x?Object.entries(x).reduce(function(P,w){var S=x2(w,2),O=S[0],C=S[1];return B(B({},P),{},ee({},O,C.scale))},{}):null}},{key:"getXScaleByAxisId",value:function(x){var P;return(P=this.state.xAxisMap)===null||P===void 0||(P=P[x])===null||P===void 0?void 0:P.scale}},{key:"getYScaleByAxisId",value:function(x){var P;return(P=this.state.yAxisMap)===null||P===void 0||(P=P[x])===null||P===void 0?void 0:P.scale}},{key:"getItemByXY",value:function(x){var P=this.state,w=P.formattedGraphicalItems,S=P.activeItem;if(w&&w.length)for(var O=0,C=w.length;O<C;O++){var _=w[O],T=_.props,A=_.item,j=A.type.defaultProps!==void 0?B(B({},A.type.defaultProps),A.props):A.props,$=Er(A.type);if($==="Bar"){var M=(T.data||[]).find(function(N){return HZ(x,N)});if(M)return{graphicalItem:_,payload:M}}else if($==="RadialBar"){var I=(T.data||[]).find(function(N){return tb(x,N)});if(I)return{graphicalItem:_,payload:I}}else if(xd(_,S)||wd(_,S)||ou(_,S)){var R=WJ({graphicalItem:_,activeTooltipItem:S,itemData:j.data}),L=j.activeIndex===void 0?R:j.activeIndex;return{graphicalItem:B(B({},_),{},{childIndex:L}),payload:ou(_,S)?j.data[R]:_.props.data[R]}}}return null}},{key:"render",value:function(){var x=this;if(!qx(this))return null;var P=this.props,w=P.children,S=P.className,O=P.width,C=P.height,_=P.style,T=P.compact,A=P.title,j=P.desc,$=w2(P,cae),M=ie($,!1);if(T)return E.createElement(Jb,{state:this.state,width:this.props.width,height:this.props.height,clipPathId:this.clipPathId},E.createElement(Jh,Jo({},M,{width:O,height:C,title:A,desc:j}),this.renderClipPath(),Kx(w,this.renderMap)));if(this.props.accessibilityLayer){var I,R;M.tabIndex=(I=this.props.tabIndex)!==null&&I!==void 0?I:0,M.role=(R=this.props.role)!==null&&R!==void 0?R:"application",M.onKeyDown=function(N){x.accessibilityManager.keyboardEvent(N)},M.onFocus=function(){x.accessibilityManager.focus()}}var L=this.parseEventsOfWrapper();return E.createElement(Jb,{state:this.state,width:this.props.width,height:this.props.height,clipPathId:this.clipPathId},E.createElement("div",Jo({className:ue("recharts-wrapper",S),style:B({position:"relative",cursor:"default",width:O,height:C},_)},L,{ref:function(D){x.container=D}}),E.createElement(Jh,Jo({},M,{width:O,height:C,title:A,desc:j,style:_ae}),this.renderClipPath(),Kx(w,this.renderMap)),this.renderLegend(),this.renderTooltip()))}}])}(k.Component);ee(h,"displayName",n),ee(h,"defaultProps",B({layout:"horizontal",stackOffset:"none",barCategoryGap:"10%",barGap:4,margin:{top:5,right:5,bottom:5,left:5},reverseStackOrder:!1,syncMethod:"index"},c)),ee(h,"getDerivedStateFromProps",function(g,v){var y=g.dataKey,x=g.data,P=g.children,w=g.width,S=g.height,O=g.layout,C=g.stackOffset,_=g.margin,T=v.dataStartIndex,A=v.dataEndIndex;if(v.updateId===void 0){var j=S2(g);return B(B(B({},j),{},{updateId:0},m(B(B({props:g},j),{},{updateId:0}),v)),{},{prevDataKey:y,prevData:x,prevWidth:w,prevHeight:S,prevLayout:O,prevStackOffset:C,prevMargin:_,prevChildren:P})}if(y!==v.prevDataKey||x!==v.prevData||w!==v.prevWidth||S!==v.prevHeight||O!==v.prevLayout||C!==v.prevStackOffset||!ui(_,v.prevMargin)){var $=S2(g),M={chartX:v.chartX,chartY:v.chartY,isTooltipActive:v.isTooltipActive},I=B(B({},P2(v,x,O)),{},{updateId:v.updateId+1}),R=B(B(B({},$),M),I);return B(B(B({},R),m(B({props:g},R),v)),{},{prevDataKey:y,prevData:x,prevWidth:w,prevHeight:S,prevLayout:O,prevStackOffset:C,prevMargin:_,prevChildren:P})}if(!Qh(P,v.prevChildren)){var L,N,D,H,W=Yt(P,Gi),G=W&&(L=(N=W.props)===null||N===void 0?void 0:N.startIndex)!==null&&L!==void 0?L:T,Z=W&&(D=(H=W.props)===null||H===void 0?void 0:H.endIndex)!==null&&D!==void 0?D:A,re=G!==T||Z!==A,ve=!le(x),be=ve&&!re?v.updateId:v.updateId+1;return B(B({updateId:be},m(B(B({props:g},v),{},{updateId:be,dataStartIndex:G,dataEndIndex:Z}),v)),{},{prevChildren:P,dataStartIndex:G,dataEndIndex:Z})}return null}),ee(h,"renderActiveDot",function(g,v,y){var x;return k.isValidElement(g)?x=k.cloneElement(g,v):ae(g)?x=g(v):x=E.createElement(gd,v),E.createElement(Ne,{className:"recharts-active-dot",key:y},x)});var d=k.forwardRef(function(v,y){return E.createElement(h,Jo({},v,{ref:y}))});return d.displayName=h.displayName,d},Bg=Fg({chartName:"LineChart",GraphicalChild:Bt,axisComponents:[{axisType:"xAxis",AxisComp:Mt},{axisType:"yAxis",AxisComp:_t}],formatAxisMap:Eg}),SS=Fg({chartName:"BarChart",GraphicalChild:Da,defaultTooltipEventType:"axis",validateTooltipEventTypes:["axis","item"],axisComponents:[{axisType:"xAxis",AxisComp:Mt},{axisType:"yAxis",AxisComp:_t}],formatAxisMap:Eg}),fu=Fg({chartName:"AreaChart",GraphicalChild:ht,axisComponents:[{axisType:"xAxis",AxisComp:Mt},{axisType:"yAxis",AxisComp:_t}],formatAxisMap:Eg});const Dae={light:"",dark:".dark"},OS=k.createContext(null);function kS(){const e=k.useContext(OS);if(!e)throw new Error("useChart must be used within a <ChartContainer />");return e}const Ln=k.forwardRef(({id:e,className:t,children:n,config:r,...a},o)=>{const i=k.useId(),s=`chart-${e||i.replace(/:/g,"")}`;return b.jsx(OS.Provider,{value:{config:r},children:b.jsxs("div",{"data-chart":s,ref:o,className:Se("flex aspect-video justify-center text-xs [&_.recharts-cartesian-axis-tick_text]:fill-muted-foreground [&_.recharts-cartesian-grid_line[stroke='#ccc']]:stroke-border/50 [&_.recharts-curve.recharts-tooltip-cursor]:stroke-border [&_.recharts-dot[stroke='#fff']]:stroke-transparent [&_.recharts-layer]:outline-none [&_.recharts-polar-grid_[stroke='#ccc']]:stroke-border [&_.recharts-radial-bar-background-sector]:fill-muted [&_.recharts-rectangle.recharts-tooltip-cursor]:fill-muted [&_.recharts-reference-line_[stroke='#ccc']]:stroke-border [&_.recharts-sector[stroke='#fff']]:stroke-transparent [&_.recharts-sector]:outline-none [&_.recharts-surface]:outline-none",t),...a,children:[b.jsx(Lae,{id:s,config:r}),b.jsx(JU,{children:n})]})})});Ln.displayName="Chart";const Lae=({id:e,config:t})=>{const n=Object.entries(t).filter(([r,a])=>a.theme||a.color);return n.length?b.jsx("style",{dangerouslySetInnerHTML:{__html:Object.entries(Dae).map(([r,a])=>`
${a} [data-chart=${e}] {
${n.map(([o,i])=>{var l;const s=((l=i.theme)==null?void 0:l[r])||i.color;return s?`  --color-${o}: ${s};`:null}).join(`
`)}
}
`).join(`
`)}}):null},or=Yn,Fn=k.forwardRef(({active:e,payload:t,className:n,indicator:r="dot",hideLabel:a=!1,hideIndicator:o=!1,label:i,labelFormatter:s,labelClassName:l,formatter:u,color:p,nameKey:c,labelKey:f},m)=>{const{config:h}=kS(),d=k.useMemo(()=>{var w;if(a||!(t!=null&&t.length))return null;const[v]=t,y=`${f||v.dataKey||v.name||"value"}`,x=wv(h,v,y),P=!f&&typeof i=="string"?((w=h[i])==null?void 0:w.label)||i:x==null?void 0:x.label;return s?b.jsx("div",{className:Se("font-medium",l),children:s(P,t)}):P?b.jsx("div",{className:Se("font-medium",l),children:P}):null},[i,s,t,a,l,h,f]);if(!e||!(t!=null&&t.length))return null;const g=t.length===1&&r!=="dot";return b.jsxs("div",{ref:m,className:Se("grid min-w-[8rem] items-start gap-1.5 rounded-lg border border-border/50 bg-background px-2.5 py-1.5 text-xs shadow-xl",n),children:[g?null:d,b.jsx("div",{className:"grid gap-1.5",children:t.map((v,y)=>{const x=`${c||v.name||v.dataKey||"value"}`,P=wv(h,v,x),w=p||v.payload.fill||v.color;return b.jsx("div",{className:Se("flex w-full flex-wrap items-stretch gap-2 [&>svg]:h-2.5 [&>svg]:w-2.5 [&>svg]:text-muted-foreground",r==="dot"&&"items-center"),children:u&&(v==null?void 0:v.value)!==void 0&&v.name?u(v.value,v.name,v,y,v.payload):b.jsxs(b.Fragment,{children:[P!=null&&P.icon?b.jsx(P.icon,{}):!o&&b.jsx("div",{className:Se("shrink-0 rounded-[2px] border-[--color-border] bg-[--color-bg]",{"h-2.5 w-2.5":r==="dot","w-1":r==="line","w-0 border-[1.5px] border-dashed bg-transparent":r==="dashed","my-0.5":g&&r==="dashed"}),style:{"--color-bg":w,"--color-border":w}}),b.jsxs("div",{className:Se("flex flex-1 justify-between leading-none",g?"items-end":"items-center"),children:[b.jsxs("div",{className:"grid gap-1.5",children:[g?d:null,b.jsx("span",{className:"text-muted-foreground",children:(P==null?void 0:P.label)||v.name})]}),v.value&&b.jsx("span",{className:"font-mono font-medium tabular-nums text-foreground",children:v.value.toLocaleString()})]})]})},v.dataKey)})})]})});Fn.displayName="ChartTooltip";const Fae=k.forwardRef(({className:e,hideIcon:t=!1,payload:n,verticalAlign:r="bottom",nameKey:a},o)=>{const{config:i}=kS();return n!=null&&n.length?b.jsx("div",{ref:o,className:Se("flex items-center justify-center gap-4",r==="top"?"pb-3":"pt-3",e),children:n.map(s=>{const l=`${a||s.dataKey||"value"}`,u=wv(i,s,l);return b.jsxs("div",{className:Se("flex items-center gap-1.5 [&>svg]:h-3 [&>svg]:w-3 [&>svg]:text-muted-foreground"),children:[u!=null&&u.icon&&!t?b.jsx(u.icon,{}):b.jsx("div",{className:"h-2 w-2 shrink-0 rounded-[2px]",style:{backgroundColor:s.color}}),u==null?void 0:u.label]},s.value)})}):null});Fae.displayName="ChartLegend";function wv(e,t,n){if(typeof t!="object"||t===null)return;const r="payload"in t&&typeof t.payload=="object"&&t.payload!==null?t.payload:void 0;let a=n;return n in t&&typeof t[n]=="string"?a=t[n]:r&&n in r&&typeof r[n]=="string"&&(a=r[n]),a in e?e[a]:e[n]}const Bae={empty:{icon:E4,title:"No Data Available",description:"There is no data to display at the moment."},error:{icon:VE,title:"Failed to Load Data",description:"An error occurred while loading the data. Please try again."},loading:{icon:j4,title:"Loading Data",description:"Please wait while we fetch the latest information."},"no-data":{icon:k4,title:"No Chart Data",description:"No data available to display in the chart."}},zae=({type:e="empty",title:t,description:n,action:r,className:a})=>{const o=Bae[e],i=o.icon;return b.jsx(ge,{className:`border-dashed ${a}`,children:b.jsxs(we,{className:"flex flex-col items-center justify-center py-12 px-6 text-center",children:[b.jsx("div",{className:"rounded-full bg-gray-100 p-3 mb-4",children:b.jsx(i,{className:`h-8 w-8 text-gray-400 ${e==="loading"?"animate-spin":""}`})}),b.jsx(xe,{className:"text-lg font-medium text-gray-900 mb-2",children:t||o.title}),b.jsx("p",{className:"text-sm text-gray-500 mb-6 max-w-sm",children:n||o.description}),r&&b.jsxs(dp,{onClick:r.onClick,variant:"outline",size:"sm",children:[b.jsx(j4,{className:"h-4 w-4 mr-2"}),r.label]})]})})},Bn=({onRetry:e})=>b.jsx(zae,{type:"no-data",title:"No Chart Data",description:"Unable to generate chart with the current data set.",action:e?{label:"Retry",onClick:e}:void 0}),ir=({text:e="Kind Cluster / GitHub CI",position:t="top-center",className:n=""})=>{const r=()=>{switch(t){case"bottom-left":return"bottom-8 left-4";case"bottom-right":return"bottom-8 right-4";case"bottom-center":return"bottom-8 left-1/2 transform -translate-x-1/2";case"top-left":return"top-4 left-4";case"top-right":return"top-4 right-4";case"top-center":return"left-1/2 transform -translate-x-1/2 -translate-y-1/2";default:return"bottom-8 left-1/2 transform -translate-x-1/2"}};return b.jsxs("div",{className:`absolute ${r()} z-10 flex items-center gap-1 px-2 py-1 bg-indigo-600/40 rounded-md border border-gray-200/30 shadow-sm ${n}`,children:[b.jsx(T4,{className:"h-3 w-3 text-white"}),b.jsx("span",{className:"text-xs text-white font-medium",children:e})]})},Hae=({performanceMatrix:e,benchmarkResults:t,testConfiguration:n,performanceSummary:r,latencyPercentileComparison:a})=>{const o=e.filter(c=>c.phase==="scaling-up").map(c=>({routes:c.routes,throughput:Math.round(c.throughput),latency:Number(c.meanLatency.toFixed(1))})),i=t.filter(c=>c.phase==="scaling-up").map(c=>({routes:c.routes,gateway:Math.round(c.resources.envoyGateway.memory.mean),proxy:Math.round(c.resources.envoyProxy.memory.mean),total:Math.round(c.resources.envoyGateway.memory.mean+c.resources.envoyProxy.memory.mean)})),l=(()=>{if(a&&a.length>0){const c=a.filter(f=>f.phase==="scaling-up").sort((f,m)=>m.routes-f.routes)[0];if(c)return[{percentile:"P50",value:Number(c.p50.toFixed(1)),status:"excellent"},{percentile:"P75",value:Number(c.p75.toFixed(1)),status:"excellent"},{percentile:"P90",value:Number(c.p90.toFixed(1)),status:"good"},{percentile:"P95",value:Number(c.p95.toFixed(1)),status:"acceptable"},{percentile:"P99",value:Number(c.p99.toFixed(1)),status:"watch"}]}if(t&&t.length>0){const c=t.filter(f=>f.phase==="scaling-up").sort((f,m)=>m.routes-f.routes)[0];if(c&&c.latency&&c.latency.percentiles){const f=c.latency.percentiles;return[{percentile:"P50",value:Number((f.p50/1e3).toFixed(1)),status:"excellent"},{percentile:"P75",value:Number((f.p75/1e3).toFixed(1)),status:"excellent"},{percentile:"P90",value:Number((f.p90/1e3).toFixed(1)),status:"good"},{percentile:"P95",value:Number((f.p95/1e3).toFixed(1)),status:"acceptable"},{percentile:"P99",value:Number((f.p99/1e3).toFixed(1)),status:"watch"}]}}return null})(),u=i.map(c=>({routes:c.routes,gatewayPerRoute:Number((c.gateway/c.routes).toFixed(2)),proxyPerRoute:Number((c.proxy/c.routes).toFixed(2)),totalPerRoute:Number((c.total/c.routes).toFixed(2))})),p={throughput:{label:"Throughput",color:"#8b5cf6"},latency:{label:"Latency",color:"#6366f1"},gateway:{label:"Gateway",color:"#a855f7"},proxy:{label:"Proxy",color:"#4f46e5"},totalPerRoute:{label:"Total per Route",color:"#8b5cf6"},gatewayPerRoute:{label:"Gateway per Route",color:"#a855f7"},proxyPerRoute:{label:"Proxy per Route",color:"#4f46e5"}};return b.jsxs("div",{className:"space-y-6",children:[b.jsxs("div",{className:"grid grid-cols-1 md:grid-cols-4 gap-4",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Avg Throughput"}),b.jsx(eT,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[Math.round(r.avgThroughput)," RPS"]}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Consistent across all scales"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Mean Response Time"}),b.jsx(A4,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[(r.avgLatency/1e3).toFixed(1),"ms"]}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"End-to-end as measured by Nighthawk"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Max Routes in Test"}),b.jsx($4,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsx("div",{className:"text-2xl font-bold",children:r.maxRoutes}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Routes tested successfully"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Reliability"}),b.jsx(QE,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsx("div",{className:"text-2xl font-bold",children:"100%"}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Perfect system reliability"})]})]})]}),b.jsxs("div",{className:"grid grid-cols-1 lg:grid-cols-2 gap-6",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Throughput Consistency"}),b.jsx(en,{children:"Throughput remains stable across different route scales"})]}),b.jsx(we,{children:o&&o.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:p,children:b.jsxs(fu,{data:o,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{domain:[0,"dataMax"],tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}`}),b.jsx(or,{content:b.jsx(Fn,{formatter:(c,f)=>[`${c} RPS`,"Throughput"],labelFormatter:c=>`${c} routes`})}),b.jsx(ht,{dataKey:"throughput",type:"monotone",fill:"#8b5cf6",fillOpacity:.4,stroke:"#8b5cf6",strokeWidth:2})]})})]}):b.jsx(Bn,{})})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Memory Usage"}),b.jsx(en,{children:"Memory scaling patterns for Gateway and Proxy components"})]}),b.jsx(we,{children:i&&i.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:p,children:b.jsxs(fu,{data:i,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}MB`}),b.jsx(or,{content:b.jsx(Fn,{})}),b.jsx(ht,{dataKey:"gateway",stackId:"memory",type:"monotone",fill:"#a855f7",stroke:"#a855f7"}),b.jsx(ht,{dataKey:"proxy",stackId:"memory",type:"monotone",fill:"#4f46e5",stroke:"#4f46e5"})]})})]}):b.jsx(Bn,{})})]})]}),b.jsxs("div",{className:"grid grid-cols-1 lg:grid-cols-2 gap-6",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Request RTT Distribution"}),b.jsx(en,{children:"Latency percentiles at 1000 routes (worst case)"})]}),b.jsx(we,{children:l?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:p,children:b.jsxs(SS,{data:l,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"percentile",tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}ms`}),b.jsx(or,{content:b.jsx(Fn,{formatter:(c,f)=>[`${c}ms`,"Latency"]})}),b.jsx(Da,{dataKey:"value",fill:"#6366f1",radius:[4,4,0,0]})]})})]}):b.jsx(Bn,{})})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Memory Efficiency"}),b.jsx(en,{children:"Memory usage per route shows how efficiently memory scales with route count"})]}),b.jsx(we,{children:u&&u.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:p,children:b.jsxs(Bg,{data:u,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}MB`}),b.jsx(or,{content:b.jsx(Fn,{formatter:(c,f)=>[`${c}MB per route`,f==="gatewayPerRoute"?"Gateway":f==="proxyPerRoute"?"Proxy":"Total"],labelFormatter:c=>`${c} routes`})}),b.jsx(Bt,{dataKey:"totalPerRoute",type:"monotone",stroke:"#8b5cf6",strokeWidth:3,dot:{fill:"#8b5cf6",strokeWidth:2,r:4}}),b.jsx(Bt,{dataKey:"gatewayPerRoute",type:"monotone",stroke:"#a855f7",strokeWidth:2,strokeDasharray:"5 5",dot:{fill:"#a855f7",strokeWidth:2,r:3}}),b.jsx(Bt,{dataKey:"proxyPerRoute",type:"monotone",stroke:"#4f46e5",strokeWidth:2,strokeDasharray:"3 3",dot:{fill:"#4f46e5",strokeWidth:2,r:3}})]})})]}):b.jsx(Bn,{})})]})]})]})},Gae=({latencyPercentileComparison:e,benchmarkResults:t})=>{const n=e.filter(c=>c.phase==="scaling-up").map(c=>({routes:c.routes,p50:Number(c.p50.toFixed(1)),p95:Number(c.p95.toFixed(1)),p99:Number(c.p99.toFixed(1))})),a=(()=>{if(e&&e.length>0){const c=e.filter(f=>f.phase==="scaling-up").sort((f,m)=>m.routes-f.routes)[0];if(c)return[{percentile:"P50",value:Number(c.p50.toFixed(1)),category:"Median",status:"excellent"},{percentile:"P75",value:Number(c.p75.toFixed(1)),category:"75th",status:"excellent"},{percentile:"P90",value:Number(c.p90.toFixed(1)),category:"90th",status:"good"},{percentile:"P95",value:Number(c.p95.toFixed(1)),category:"95th",status:"watch"},{percentile:"P99",value:Number(c.p99.toFixed(1)),category:"99th",status:"alert"}]}if(t&&t.length>0){const c=t.filter(f=>f.phase==="scaling-up").sort((f,m)=>m.routes-f.routes)[0];if(c&&c.latency&&c.latency.percentiles){const f=c.latency.percentiles;return[{percentile:"P50",value:Number((f.p50/1e3).toFixed(1)),category:"Median",status:"excellent"},{percentile:"P75",value:Number((f.p75/1e3).toFixed(1)),category:"75th",status:"excellent"},{percentile:"P90",value:Number((f.p90/1e3).toFixed(1)),category:"90th",status:"good"},{percentile:"P95",value:Number((f.p95/1e3).toFixed(1)),category:"95th",status:"watch"},{percentile:"P99",value:Number((f.p99/1e3).toFixed(1)),category:"99th",status:"alert"}]}}return null})(),o=t.filter(c=>c.phase==="scaling-up").map(c=>({routes:c.routes,mean:Number((c.latency.mean/1e3).toFixed(1)),p95:Number((c.latency.percentiles.p95/1e3).toFixed(1)),ratio:Number((c.latency.percentiles.p95/c.latency.mean).toFixed(1))})),s=e&&e.length>0?e.filter(c=>c.phase==="scaling-up").sort((c,f)=>c.routes-f.routes).map(c=>({scale:`${c.routes} Routes`,p50:Number(c.p50.toFixed(1)),p95:Number(c.p95.toFixed(1)),p99:Number(c.p99.toFixed(1))})):null,l={p50:{label:"P50 (Median)",color:"#8b5cf6"},p95:{label:"P95",color:"#6366f1"},p99:{label:"P99",color:"#4f46e5"},mean:{label:"Mean",color:"#a855f7"},ratio:{label:"P95/Mean Ratio",color:"#7c3aed"}},p=(()=>{if(n.length===0)return{medianLatency:0,p95Latency:0,p99Latency:0,consistencyRatio:0};const c=n[n.length-1],f=o.length>0?o.reduce((m,h)=>m+h.ratio,0)/o.length:0;return{medianLatency:(c==null?void 0:c.p50)||0,p95Latency:(c==null?void 0:c.p95)||0,p99Latency:(c==null?void 0:c.p99)||0,consistencyRatio:f}})();return b.jsxs("div",{className:"space-y-6",children:[b.jsxs("div",{className:"grid grid-cols-1 md:grid-cols-4 gap-4",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Median RTT"}),b.jsx(ZE,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[p.medianLatency.toFixed(1),"ms"]}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Consistent across all scales"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"P95 RTT"}),b.jsx(XE,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[p.p95Latency.toFixed(1),"ms"]}),b.jsxs("p",{className:"text-xs text-muted-foreground",children:["95% of requests under ",p.p95Latency.toFixed(0),"ms"]})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Tail RTT"}),b.jsx(JE,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[p.p99Latency.toFixed(1),"ms"]}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"P99 RTT (1% of requests)"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Consistency"}),b.jsx(_4,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[p.consistencyRatio.toFixed(1),":1"]}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"P95/Mean ratio"})]})]})]}),b.jsxs("div",{className:"grid grid-cols-1 lg:grid-cols-2 gap-6",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Request RTT Scaling Behavior"}),b.jsx(en,{children:"How key percentiles perform as route count increases"})]}),b.jsx(we,{children:n&&n.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:l,children:b.jsxs(fu,{data:n,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}ms`}),b.jsx(or,{content:b.jsx(Fn,{labelFormatter:c=>`${c} routes`})}),b.jsx(ht,{dataKey:"p99",stackId:"latency",type:"monotone",fill:"#4f46e5",fillOpacity:.3,stroke:"#4f46e5",strokeWidth:1}),b.jsx(ht,{dataKey:"p95",stackId:"latency",type:"monotone",fill:"#6366f1",fillOpacity:.4,stroke:"#6366f1",strokeWidth:2}),b.jsx(ht,{dataKey:"p50",stackId:"latency",type:"monotone",fill:"#8b5cf6",fillOpacity:.6,stroke:"#8b5cf6",strokeWidth:3})]})})]}):b.jsx(Bn,{})})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Request RTT Distribution"}),b.jsx(en,{children:"Percentile breakdown at 1000 routes (worst case scenario)"})]}),b.jsx(we,{children:a?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:l,children:b.jsxs(SS,{data:a,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"percentile",tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}ms`}),b.jsx(or,{content:b.jsx(Fn,{formatter:(c,f)=>[`${c}ms`,"RTT"],labelFormatter:c=>`${c} Percentile`})}),b.jsx(Da,{dataKey:"value",fill:"#6366f1",radius:[4,4,0,0]})]})})]}):b.jsx(Bn,{})})]})]}),b.jsxs("div",{className:"grid grid-cols-1 lg:grid-cols-2 gap-6",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Request RTT Consistency"}),b.jsx(en,{children:"P95/Mean ratio shows how predictable RTT is"})]}),b.jsx(we,{children:o&&o.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:l,children:b.jsxs(Bg,{data:o,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{yAxisId:"ratio",orientation:"left",tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}x`}),b.jsx(_t,{yAxisId:"latency",orientation:"right",tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:c=>`${c}ms`}),b.jsx(or,{content:b.jsx(Fn,{labelFormatter:c=>`${c} routes`})}),b.jsx(Bt,{yAxisId:"latency",dataKey:"mean",type:"monotone",stroke:"#a855f7",strokeWidth:2,dot:{fill:"#a855f7",strokeWidth:2,r:3}}),b.jsx(Bt,{yAxisId:"ratio",dataKey:"ratio",type:"monotone",stroke:"#7c3aed",strokeWidth:3,strokeDasharray:"5 5",dot:{fill:"#7c3aed",strokeWidth:2,r:4}})]})})]}):b.jsx(Bn,{})})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Performance Summary"}),b.jsx(en,{children:"RTT performance at different scales"})]}),b.jsx(we,{children:s&&s.length>0?b.jsx("div",{className:"space-y-4",children:s.map((c,f)=>b.jsx("div",{className:"flex items-center justify-between p-3 bg-muted/30 rounded-lg",children:b.jsxs("div",{className:"space-y-1",children:[b.jsx("p",{className:"font-medium",children:c.scale}),b.jsxs("div",{className:"flex space-x-4 text-sm text-muted-foreground",children:[b.jsxs("span",{children:["P50: ",c.p50,"ms"]}),b.jsxs("span",{children:["P95: ",c.p95,"ms"]}),b.jsxs("span",{children:["P99: ",c.p99,"ms"]})]})]})},f))}):b.jsx(Bn,{})})]})]})]})},Uae=({resourceTrends:e,benchmarkResults:t})=>{const n=t.filter(f=>f.phase==="scaling-up").map(f=>({routes:f.routes,gateway:Math.round(f.resources.envoyGateway.memory.mean),proxy:Math.round(f.resources.envoyProxy.memory.mean),total:Math.round(f.resources.envoyGateway.memory.mean+f.resources.envoyProxy.memory.mean)})),r=t.filter(f=>f.phase==="scaling-up").map(f=>({routes:f.routes,gatewayMean:Number(f.resources.envoyGateway.cpu.mean.toFixed(1)),gatewayMax:Number(f.resources.envoyGateway.cpu.max.toFixed(1)),proxyMean:Number(f.resources.envoyProxy.cpu.mean.toFixed(1)),proxyMax:Number(f.resources.envoyProxy.cpu.max.toFixed(1))})),a=n.map(f=>({routes:f.routes,gatewayPerRoute:Number((f.gateway/f.routes).toFixed(2)),proxyPerRoute:Number((f.proxy/f.routes).toFixed(2)),totalPerRoute:Number((f.total/f.routes).toFixed(2))})),o=()=>{if(n.length===0)return[{component:"Gateway",min:0,max:0,scaling:"No Data",efficiency:"Unknown"},{component:"Proxy",min:0,max:0,scaling:"No Data",efficiency:"Unknown"},{component:"Total",min:0,max:0,scaling:"No Data",efficiency:"Unknown"}];const f=n.map(d=>d.gateway),m=n.map(d=>d.proxy),h=n.map(d=>d.total);return[{component:"Gateway",min:Math.min(...f),max:Math.max(...f),scaling:i(n.map(d=>d.routes),f),efficiency:s(n.map(d=>d.routes),f)},{component:"Proxy",min:Math.min(...m),max:Math.max(...m),scaling:i(n.map(d=>d.routes),m),efficiency:s(n.map(d=>d.routes),m)},{component:"Total",min:Math.min(...h),max:Math.max(...h),scaling:i(n.map(d=>d.routes),h),efficiency:s(n.map(d=>d.routes),h)}]},i=(f,m)=>{if(f.length<3)return"Insufficient Data";const h=f.length,d=f.reduce(($,M)=>$+M,0),g=m.reduce(($,M)=>$+M,0),v=f.reduce(($,M,I)=>$+M*m[I],0),y=f.reduce(($,M)=>$+M*M,0),x=m.reduce(($,M)=>$+M*M,0),P=h*v-d*g,w=Math.sqrt((h*y-d*d)*(h*x-g*g)),S=w===0?0:P/w,O=S*S,C=[];for(let $=1;$<m.length;$++)C.push(m[$]-m[$-1]);const _=C.reduce(($,M)=>$+M,0)/C.length,T=C.reduce(($,M)=>$+Math.pow(M-_,2),0)/C.length,A=Math.sqrt(T),j=_===0?0:Math.abs(A/_);return O>.95?"Highly Linear":O>.85?"Linear":j>1.5?"Step-wise":O>.7?"Moderately Linear":"Variable"},s=(f,m)=>{if(f.length<2)return"Unknown";const h=f.map((w,S)=>m[S]/w),d=h[0],g=h[h.length-1],v=(d-g)/d,y=h.reduce((w,S)=>w+S,0)/h.length,x=h.reduce((w,S)=>w+Math.pow(S-y,2),0)/h.length,P=Math.sqrt(x)/y;return v>.3?"Excellent":v>.1||P<.2?"Good":P<.4?"Moderate":"Variable"},l=o(),u={gateway:{label:"Gateway",color:"#a855f7"},proxy:{label:"Proxy",color:"#4f46e5"},gatewayMean:{label:"Gateway Mean",color:"#8b5cf6"},gatewayMax:{label:"Gateway Peak",color:"#c084fc"},proxyMean:{label:"Proxy Mean",color:"#6366f1"},proxyMax:{label:"Proxy Peak",color:"#818cf8"}},c=(()=>{if(n.length===0||r.length===0||a.length===0)return{gatewayMemoryRange:"0-0MB",proxyMemoryRange:"0-0MB",peakCPU:0,memoryPerRouteAtScale:0};const f=n.map(S=>S.gateway),m=n.map(S=>S.proxy),h=Math.min(...f),d=Math.max(...f),g=Math.min(...m),v=Math.max(...m),y=r.flatMap(S=>[S.gatewayMax,S.proxyMax]),x=Math.max(...y),P=a[a.length-1],w=(P==null?void 0:P.totalPerRoute)||0;return{gatewayMemoryRange:`${h}-${d}MB`,proxyMemoryRange:`${g}-${v}MB`,peakCPU:Math.round(x),memoryPerRouteAtScale:w}})();return b.jsxs("div",{className:"space-y-6",children:[b.jsxs("div",{className:"grid grid-cols-1 md:grid-cols-4 gap-4",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Gateway Memory"}),b.jsx(E4,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsx("div",{className:"text-2xl font-bold",children:c.gatewayMemoryRange}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Linear scaling pattern"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Proxy Memory"}),b.jsx(N4,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsx("div",{className:"text-2xl font-bold",children:c.proxyMemoryRange}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Efficient step-wise scaling"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Peak CPU"}),b.jsx(O4,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[c.peakCPU,"%"]}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Brief spikes, stable avg"})]})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{className:"flex flex-row items-center justify-between space-y-0 pb-2",children:[b.jsx(xe,{className:"text-sm font-medium",children:"Efficiency"}),b.jsx($4,{className:"h-4 w-4 text-muted-foreground"})]}),b.jsxs(we,{children:[b.jsxs("div",{className:"text-2xl font-bold",children:[c.memoryPerRouteAtScale.toFixed(2),"MB"]}),b.jsx("p",{className:"text-xs text-muted-foreground",children:"Memory per route at scale"})]})]})]}),b.jsxs("div",{className:"grid grid-cols-1 lg:grid-cols-2 gap-6",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Memory Scaling"}),b.jsx(en,{children:"How Gateway and Proxy memory usage grows with route count"})]}),b.jsx(we,{children:n&&n.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:u,children:b.jsxs(fu,{data:n,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:f=>`${f}MB`}),b.jsx(or,{content:b.jsx(Fn,{labelFormatter:f=>`${f} routes`})}),b.jsx(ht,{dataKey:"gateway",stackId:"memory",type:"monotone",fill:"#a855f7",fillOpacity:.6,stroke:"#a855f7",strokeWidth:2}),b.jsx(ht,{dataKey:"proxy",stackId:"memory",type:"monotone",fill:"#4f46e5",fillOpacity:.6,stroke:"#4f46e5",strokeWidth:2})]})})]}):b.jsx(Bn,{})})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"CPU Usage Patterns"}),b.jsx(en,{children:"Mean vs peak CPU usage showing burst characteristics"})]}),b.jsx(we,{children:r&&r.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:u,children:b.jsxs(fu,{data:r,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:f=>`${f}%`}),b.jsx(or,{content:b.jsx(Fn,{labelFormatter:f=>`${f} routes`})}),b.jsx(ht,{dataKey:"gatewayMax",type:"monotone",fill:"#c084fc",fillOpacity:.3,stroke:"#c084fc",strokeWidth:1}),b.jsx(ht,{dataKey:"proxyMax",type:"monotone",fill:"#818cf8",fillOpacity:.3,stroke:"#818cf8",strokeWidth:1}),b.jsx(Bt,{dataKey:"gatewayMean",type:"monotone",stroke:"#8b5cf6",strokeWidth:3,dot:{fill:"#8b5cf6",strokeWidth:2,r:3}}),b.jsx(Bt,{dataKey:"proxyMean",type:"monotone",stroke:"#6366f1",strokeWidth:3,dot:{fill:"#6366f1",strokeWidth:2,r:3}})]})})]}):b.jsx(Bn,{})})]})]}),b.jsxs("div",{className:"grid grid-cols-1 lg:grid-cols-2 gap-6",children:[b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Memory Efficiency"}),b.jsx(en,{children:"Memory usage per route shows how efficiently memory scales with route count"})]}),b.jsx(we,{children:a&&a.length>0?b.jsxs("div",{className:"relative",children:[b.jsx(ir,{}),b.jsx(Ln,{config:u,children:b.jsxs(Bg,{data:a,children:[b.jsx(Dn,{strokeDasharray:"3 3"}),b.jsx(Mt,{dataKey:"routes",type:"number",scale:"linear",domain:[0,1100],ticks:[10,50,100,300,500,1e3],tickLine:!1,axisLine:!1,tickMargin:8}),b.jsx(_t,{tickLine:!1,axisLine:!1,tickMargin:8,tickFormatter:f=>`${f}MB`}),b.jsx(or,{content:b.jsx(Fn,{formatter:(f,m)=>[`${f}MB per route`,m==="gatewayPerRoute"?"Gateway":m==="proxyPerRoute"?"Proxy":"Total"],labelFormatter:f=>`${f} routes`})}),b.jsx(Bt,{dataKey:"totalPerRoute",type:"monotone",stroke:"#8b5cf6",strokeWidth:3,dot:{fill:"#8b5cf6",strokeWidth:2,r:4}}),b.jsx(Bt,{dataKey:"gatewayPerRoute",type:"monotone",stroke:"#a855f7",strokeWidth:2,strokeDasharray:"5 5",dot:{fill:"#a855f7",strokeWidth:2,r:3}}),b.jsx(Bt,{dataKey:"proxyPerRoute",type:"monotone",stroke:"#4f46e5",strokeWidth:2,strokeDasharray:"3 3",dot:{fill:"#4f46e5",strokeWidth:2,r:3}})]})})]}):b.jsx(Bn,{})})]}),b.jsxs(ge,{children:[b.jsxs(Pe,{children:[b.jsx(xe,{children:"Memory Scaling Summary"}),b.jsx(en,{children:"Memory usage characteristics and scaling patterns across different components"})]}),b.jsx(we,{children:b.jsx("div",{className:"space-y-4",children:l.map((f,m)=>b.jsxs("div",{className:"flex items-center justify-between p-3 bg-muted/30 rounded-lg",children:[b.jsxs("div",{className:"space-y-1",children:[b.jsx("p",{className:"font-medium",children:f.component}),b.jsxs("div",{className:"flex space-x-4 text-sm text-muted-foreground",children:[b.jsxs("span",{children:["Range: ",f.min,"-",f.max,"MB"]}),b.jsxs("span",{children:["Pattern: ",f.scaling]})]})]}),b.jsxs("div",{className:"flex items-center space-x-2",children:[b.jsx(_4,{className:"h-4 w-4 text-green-600"}),b.jsx("span",{className:"text-sm font-medium text-green-700",children:f.efficiency})]})]},m))})})]})]})]})};var Um="rovingFocusGroup.onEntryFocus",Wae={bubbles:!1,cancelable:!0},ju="RovingFocusGroup",[bv,CS,qae]=n6(ju),[Vae,_S]=is(ju,[qae]),[Kae,Xae]=Vae(ju),AS=k.forwardRef((e,t)=>b.jsx(bv.Provider,{scope:e.__scopeRovingFocusGroup,children:b.jsx(bv.Slot,{scope:e.__scopeRovingFocusGroup,children:b.jsx(Yae,{...e,ref:t})})}));AS.displayName=ju;var Yae=k.forwardRef((e,t)=>{const{__scopeRovingFocusGroup:n,orientation:r,loop:a=!1,dir:o,currentTabStopId:i,defaultCurrentTabStopId:s,onCurrentTabStopIdChange:l,onEntryFocus:u,preventScrollOnEntryFocus:p=!1,...c}=e,f=k.useRef(null),m=Xe(t,f),h=xy(o),[d,g]=pp({prop:i,defaultProp:s??null,onChange:l,caller:ju}),[v,y]=k.useState(!1),x=Oa(u),P=CS(n),w=k.useRef(!1),[S,O]=k.useState(0);return k.useEffect(()=>{const C=f.current;if(C)return C.addEventListener(Um,x),()=>C.removeEventListener(Um,x)},[x]),b.jsx(Kae,{scope:n,orientation:r,dir:h,loop:a,currentTabStopId:d,onItemFocus:k.useCallback(C=>g(C),[g]),onItemShiftTab:k.useCallback(()=>y(!0),[]),onFocusableItemAdd:k.useCallback(()=>O(C=>C+1),[]),onFocusableItemRemove:k.useCallback(()=>O(C=>C-1),[]),children:b.jsx(Ae.div,{tabIndex:v||S===0?-1:0,"data-orientation":r,...c,ref:m,style:{outline:"none",...e.style},onMouseDown:fe(e.onMouseDown,()=>{w.current=!0}),onFocus:fe(e.onFocus,C=>{const _=!w.current;if(C.target===C.currentTarget&&_&&!v){const T=new CustomEvent(Um,Wae);if(C.currentTarget.dispatchEvent(T),!T.defaultPrevented){const A=P().filter(R=>R.focusable),j=A.find(R=>R.active),$=A.find(R=>R.id===d),I=[j,$,...A].filter(Boolean).map(R=>R.ref.current);jS(I,p)}}w.current=!1}),onBlur:fe(e.onBlur,()=>y(!1))})})}),ES="RovingFocusGroupItem",TS=k.forwardRef((e,t)=>{const{__scopeRovingFocusGroup:n,focusable:r=!0,active:a=!1,tabStopId:o,children:i,...s}=e,l=Pu(),u=o||l,p=Xae(ES,n),c=p.currentTabStopId===u,f=CS(n),{onFocusableItemAdd:m,onFocusableItemRemove:h,currentTabStopId:d}=p;return k.useEffect(()=>{if(r)return m(),()=>h()},[r,m,h]),b.jsx(bv.ItemSlot,{scope:n,id:u,focusable:r,active:a,children:b.jsx(Ae.span,{tabIndex:c?0:-1,"data-orientation":p.orientation,...s,ref:t,onMouseDown:fe(e.onMouseDown,g=>{r?p.onItemFocus(u):g.preventDefault()}),onFocus:fe(e.onFocus,()=>p.onItemFocus(u)),onKeyDown:fe(e.onKeyDown,g=>{if(g.key==="Tab"&&g.shiftKey){p.onItemShiftTab();return}if(g.target!==g.currentTarget)return;const v=Jae(g,p.orientation,p.dir);if(v!==void 0){if(g.metaKey||g.ctrlKey||g.altKey||g.shiftKey)return;g.preventDefault();let x=f().filter(P=>P.focusable).map(P=>P.ref.current);if(v==="last")x.reverse();else if(v==="prev"||v==="next"){v==="prev"&&x.reverse();const P=x.indexOf(g.currentTarget);x=p.loop?eoe(x,P+1):x.slice(P+1)}setTimeout(()=>jS(x))}}),children:typeof i=="function"?i({isCurrentTabStop:c,hasTabStop:d!=null}):i})})});TS.displayName=ES;var Qae={ArrowLeft:"prev",ArrowUp:"prev",ArrowRight:"next",ArrowDown:"next",PageUp:"first",Home:"first",PageDown:"last",End:"last"};function Zae(e,t){return t!=="rtl"?e:e==="ArrowLeft"?"ArrowRight":e==="ArrowRight"?"ArrowLeft":e}function Jae(e,t,n){const r=Zae(e.key,n);if(!(t==="vertical"&&["ArrowLeft","ArrowRight"].includes(r))&&!(t==="horizontal"&&["ArrowUp","ArrowDown"].includes(r)))return Qae[r]}function jS(e,t=!1){const n=document.activeElement;for(const r of e)if(r===n||(r.focus({preventScroll:t}),document.activeElement!==n))return}function eoe(e,t){return e.map((n,r)=>e[(t+r)%e.length])}var toe=AS,noe=TS;function roe(e,t){return k.useReducer((n,r)=>t[n][r]??n,e)}var zg=e=>{const{present:t,children:n}=e,r=aoe(t),a=typeof n=="function"?n({present:r.isPresent}):k.Children.only(n),o=Xe(r.ref,ooe(a));return typeof n=="function"||r.isPresent?k.cloneElement(a,{ref:o}):null};zg.displayName="Presence";function aoe(e){const[t,n]=k.useState(),r=k.useRef(null),a=k.useRef(e),o=k.useRef("none"),i=e?"mounted":"unmounted",[s,l]=roe(i,{mounted:{UNMOUNT:"unmounted",ANIMATION_OUT:"unmountSuspended"},unmountSuspended:{MOUNT:"mounted",ANIMATION_END:"unmounted"},unmounted:{MOUNT:"mounted"}});return k.useEffect(()=>{const u=yc(r.current);o.current=s==="mounted"?u:"none"},[s]),Et(()=>{const u=r.current,p=a.current;if(p!==e){const f=o.current,m=yc(u);e?l("MOUNT"):m==="none"||(u==null?void 0:u.display)==="none"?l("UNMOUNT"):l(p&&f!==m?"ANIMATION_OUT":"UNMOUNT"),a.current=e}},[e,l]),Et(()=>{if(t){let u;const p=t.ownerDocument.defaultView??window,c=m=>{const d=yc(r.current).includes(m.animationName);if(m.target===t&&d&&(l("ANIMATION_END"),!a.current)){const g=t.style.animationFillMode;t.style.animationFillMode="forwards",u=p.setTimeout(()=>{t.style.animationFillMode==="forwards"&&(t.style.animationFillMode=g)})}},f=m=>{m.target===t&&(o.current=yc(r.current))};return t.addEventListener("animationstart",f),t.addEventListener("animationcancel",c),t.addEventListener("animationend",c),()=>{p.clearTimeout(u),t.removeEventListener("animationstart",f),t.removeEventListener("animationcancel",c),t.removeEventListener("animationend",c)}}else l("ANIMATION_END")},[t,l]),{isPresent:["mounted","unmountSuspended"].includes(s),ref:k.useCallback(u=>{r.current=u?getComputedStyle(u):null,n(u)},[])}}function yc(e){return(e==null?void 0:e.animationName)||"none"}function ooe(e){var r,a;let t=(r=Object.getOwnPropertyDescriptor(e.props,"ref"))==null?void 0:r.get,n=t&&"isReactWarning"in t&&t.isReactWarning;return n?e.ref:(t=(a=Object.getOwnPropertyDescriptor(e,"ref"))==null?void 0:a.get,n=t&&"isReactWarning"in t&&t.isReactWarning,n?e.props.ref:e.props.ref||e.ref)}var _d="Tabs",[ioe,_ie]=is(_d,[_S]),NS=_S(),[soe,Hg]=ioe(_d),$S=k.forwardRef((e,t)=>{const{__scopeTabs:n,value:r,onValueChange:a,defaultValue:o,orientation:i="horizontal",dir:s,activationMode:l="automatic",...u}=e,p=xy(s),[c,f]=pp({prop:r,onChange:a,defaultProp:o??"",caller:_d});return b.jsx(soe,{scope:n,baseId:Pu(),value:c,onValueChange:f,orientation:i,dir:p,activationMode:l,children:b.jsx(Ae.div,{dir:p,"data-orientation":i,...u,ref:t})})});$S.displayName=_d;var MS="TabsList",RS=k.forwardRef((e,t)=>{const{__scopeTabs:n,loop:r=!0,...a}=e,o=Hg(MS,n),i=NS(n);return b.jsx(toe,{asChild:!0,...i,orientation:o.orientation,dir:o.dir,loop:r,children:b.jsx(Ae.div,{role:"tablist","aria-orientation":o.orientation,...a,ref:t})})});RS.displayName=MS;var IS="TabsTrigger",DS=k.forwardRef((e,t)=>{const{__scopeTabs:n,value:r,disabled:a=!1,...o}=e,i=Hg(IS,n),s=NS(n),l=BS(i.baseId,r),u=zS(i.baseId,r),p=r===i.value;return b.jsx(noe,{asChild:!0,...s,focusable:!a,active:p,children:b.jsx(Ae.button,{type:"button",role:"tab","aria-selected":p,"aria-controls":u,"data-state":p?"active":"inactive","data-disabled":a?"":void 0,disabled:a,id:l,...o,ref:t,onMouseDown:fe(e.onMouseDown,c=>{!a&&c.button===0&&c.ctrlKey===!1?i.onValueChange(r):c.preventDefault()}),onKeyDown:fe(e.onKeyDown,c=>{[" ","Enter"].includes(c.key)&&i.onValueChange(r)}),onFocus:fe(e.onFocus,()=>{const c=i.activationMode!=="manual";!p&&!a&&c&&i.onValueChange(r)})})})});DS.displayName=IS;var LS="TabsContent",FS=k.forwardRef((e,t)=>{const{__scopeTabs:n,value:r,forceMount:a,children:o,...i}=e,s=Hg(LS,n),l=BS(s.baseId,r),u=zS(s.baseId,r),p=r===s.value,c=k.useRef(p);return k.useEffect(()=>{const f=requestAnimationFrame(()=>c.current=!1);return()=>cancelAnimationFrame(f)},[]),b.jsx(zg,{present:a||p,children:({present:f})=>b.jsx(Ae.div,{"data-state":p?"active":"inactive","data-orientation":s.orientation,role:"tabpanel","aria-labelledby":l,hidden:!f,id:u,tabIndex:0,...i,ref:t,style:{...e.style,animationDuration:c.current?"0s":void 0},children:f&&o})})});FS.displayName=LS;function BS(e,t){return`${e}-trigger-${t}`}function zS(e,t){return`${e}-content-${t}`}var loe=$S,HS=RS,GS=DS,US=FS;const uoe=loe,WS=k.forwardRef(({className:e,...t},n)=>b.jsx(HS,{ref:n,className:Se("inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground",e),...t}));WS.displayName=HS.displayName;const Mc=k.forwardRef(({className:e,...t},n)=>b.jsx(GS,{ref:n,className:Se("inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm",e),...t}));Mc.displayName=GS.displayName;const Rc=k.forwardRef(({className:e,...t},n)=>b.jsx(US,{ref:n,className:Se("mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",e),...t}));Rc.displayName=US.displayName;const coe={metadata:{version:"1.3.3",runId:"1.3.3-1750189826778",date:"2025-05-09",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.3.3",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.3.3/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5302.05,totalRequests:159067,latency:{min:384,mean:7440,max:68399,pstdev:12191,percentiles:{p50:3692,p75:5903,p80:6670,p90:10777,p95:47519,p99:54405,p999:59346}},resources:{envoyGateway:{memory:{min:126.17,max:126.17,mean:126.17},cpu:{min:.95,max:.95,mean:.95}},envoyProxy:{memory:{min:25.93,max:25.93,mean:25.93},cpu:{min:30.42,max:30.42,mean:30.42}}},poolOverflow:358,upstreamConnections:42},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5270.25,totalRequests:158112,latency:{min:392,mean:7067,max:77467,pstdev:12552,percentiles:{p50:3330,p75:5225,p80:5887,p90:8906,p95:49113,p99:55103,p999:60295}},resources:{envoyGateway:{memory:{min:145.89,max:145.89,mean:145.89},cpu:{min:1.66,max:1.66,mean:1.66}},envoyProxy:{memory:{min:32.1,max:32.1,mean:32.1},cpu:{min:60.71,max:60.71,mean:60.71}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5307.61,totalRequests:159230,latency:{min:387,mean:7350,max:76185,pstdev:12732,percentiles:{p50:3508,p75:5524,p80:6220,p90:9652,p95:49100,p99:55363,p999:62654}},resources:{envoyGateway:{memory:{min:153.05,max:153.05,mean:153.05},cpu:{min:3.08,max:3.08,mean:3.08}},envoyProxy:{memory:{min:38.29,max:38.29,mean:38.29},cpu:{min:91.5,max:91.5,mean:91.5}}},poolOverflow:359,upstreamConnections:41},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5139.63,totalRequests:154189,latency:{min:390,mean:7376,max:78098,pstdev:12993,percentiles:{p50:3429,p75:5431,p80:6116,p90:9633,p95:49809,p99:56043,p999:64651}},resources:{envoyGateway:{memory:{min:152.28,max:152.28,mean:152.28},cpu:{min:15.6,max:15.6,mean:15.6}},envoyProxy:{memory:{min:56.33,max:56.33,mean:56.33},cpu:{min:124.15,max:124.15,mean:124.15}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5169.8,totalRequests:155094,latency:{min:374,mean:7385,max:105918,pstdev:13085,percentiles:{p50:3382,p75:5336,p80:6024,p90:9658,p95:50104,p99:56487,p999:66418}},resources:{envoyGateway:{memory:{min:166.29,max:166.29,mean:166.29},cpu:{min:28.13,max:28.13,mean:28.13}},envoyProxy:{memory:{min:76.53,max:76.53,mean:76.53},cpu:{min:157.45,max:157.45,mean:157.45}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:4963.46,totalRequests:148908,latency:{min:374,mean:7335,max:94580,pstdev:13294,percentiles:{p50:3255,p75:5279,p80:5998,p90:9532,p95:50294,p99:58908,p999:70729}},resources:{envoyGateway:{memory:{min:199.79,max:199.79,mean:199.79},cpu:{min:61.67,max:61.67,mean:61.67}},envoyProxy:{memory:{min:120.77,max:120.77,mean:120.77},cpu:{min:195.1,max:195.1,mean:195.1}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5200.07,totalRequests:156001,latency:{min:395,mean:7530,max:81936,pstdev:13142,percentiles:{p50:3438,p75:5497,p80:6226,p90:10218,p95:50345,p99:56436,p999:61919}},resources:{envoyGateway:{memory:{min:221.87,max:221.87,mean:221.87},cpu:{min:114.77,max:114.77,mean:114.77}},envoyProxy:{memory:{min:121.27,max:121.27,mean:121.27},cpu:{min:356.4,max:356.4,mean:356.4}}},poolOverflow:359,upstreamConnections:41},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5229.66,totalRequests:156893,latency:{min:368,mean:6911,max:72273,pstdev:12396,percentiles:{p50:3227,p75:5075,p80:5695,p90:8751,p95:48676,p99:54822,p999:59865}},resources:{envoyGateway:{memory:{min:221.8,max:221.8,mean:221.8},cpu:{min:114.13,max:114.13,mean:114.13}},envoyProxy:{memory:{min:121.22,max:121.22,mean:121.22},cpu:{min:325.9,max:325.9,mean:325.9}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5099.34,totalRequests:152980,latency:{min:357,mean:6637,max:68100,pstdev:12012,percentiles:{p50:3167,p75:4972,p80:5580,p90:8302,p95:47650,p99:53993,p999:59299}},resources:{envoyGateway:{memory:{min:182.69,max:182.69,mean:182.69},cpu:{min:113.02,max:113.02,mean:113.02}},envoyProxy:{memory:{min:121.2,max:121.2,mean:121.2},cpu:{min:295.3,max:295.3,mean:295.3}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5145.13,totalRequests:154357,latency:{min:379,mean:7234,max:111517,pstdev:12814,percentiles:{p50:3381,p75:5331,p80:5993,p90:9269,p95:49465,p99:55562,p999:64862}},resources:{envoyGateway:{memory:{min:159.09,max:159.09,mean:159.09},cpu:{min:102.57,max:102.57,mean:102.57}},envoyProxy:{memory:{min:121.08,max:121.08,mean:121.08},cpu:{min:263.29,max:263.29,mean:263.29}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5130.95,totalRequests:153929,latency:{min:365,mean:7098,max:95834,pstdev:12843,percentiles:{p50:3267,p75:5110,p80:5752,p90:8975,p95:49702,p99:56408,p999:67330}},resources:{envoyGateway:{memory:{min:163.91,max:163.91,mean:163.91},cpu:{min:91.18,max:91.18,mean:91.18}},envoyProxy:{memory:{min:120.91,max:120.91,mean:120.91},cpu:{min:230.77,max:230.77,mean:230.77}}},poolOverflow:362,upstreamConnections:38}]},poe={metadata:{version:"1.4.0",runId:"1.4.0-1750189826781",date:"2025-05-14",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.4.0",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.4.0/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scaling up httproutes to 10 with 2 routes per hostname",routes:10,routesPerHostname:2,phase:"scaling-up",throughput:5433.23,totalRequests:163e3,latency:{min:367,mean:6485,max:65673,pstdev:11042,percentiles:{p50:3222,p75:5166,p80:5879,p90:9253,p95:43970,p99:52201,p999:56854}},resources:{envoyGateway:{memory:{min:128.52,max:156.24,mean:150.28},cpu:{min:.13,max:.67,mean:.44}},envoyProxy:{memory:{min:0,max:27.01,mean:21.73},cpu:{min:0,max:72.36,mean:2.32}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling up httproutes to 50 with 10 routes per hostname",routes:50,routesPerHostname:10,phase:"scaling-up",throughput:5340.69,totalRequests:160221,latency:{min:355,mean:6582,max:88674,pstdev:11849,percentiles:{p50:3134,p75:4984,p80:5597,p90:8545,p95:46972,p99:53663,p999:59625}},resources:{envoyGateway:{memory:{min:155.3,max:169.25,mean:159.23},cpu:{min:.27,max:4.73,mean:.91}},envoyProxy:{memory:{min:26.64,max:33.18,mean:32.22},cpu:{min:0,max:99.91,mean:11.1}}},poolOverflow:363,upstreamConnections:37},{testName:"scaling up httproutes to 100 with 20 routes per hostname",routes:100,routesPerHostname:20,phase:"scaling-up",throughput:5205.41,totalRequests:156162,latency:{min:384,mean:4221,max:63500,pstdev:9204,percentiles:{p50:2039,p75:3063,p80:3399,p90:4605,p95:7844,p99:49758,p999:54032}},resources:{envoyGateway:{memory:{min:156.77,max:174.36,mean:168.53},cpu:{min:.4,max:8.8,mean:1.37}},envoyProxy:{memory:{min:32.83,max:37.05,mean:36.58},cpu:{min:0,max:99.97,mean:4.39}}},poolOverflow:377,upstreamConnections:23},{testName:"scaling up httproutes to 300 with 60 routes per hostname",routes:300,routesPerHostname:60,phase:"scaling-up",throughput:5271.84,totalRequests:158157,latency:{min:393,mean:7209,max:93327,pstdev:12640,percentiles:{p50:3281,p75:5443,p80:6241,p90:10258,p95:48494,p99:55793,p999:65497}},resources:{envoyGateway:{memory:{min:179.66,max:187.8,mean:184.81},cpu:{min:.53,max:27.13,mean:4.87}},envoyProxy:{memory:{min:53.06,max:57.31,mean:56.58},cpu:{min:0,max:99.96,mean:15.82}}},poolOverflow:360,upstreamConnections:40},{testName:"scaling up httproutes to 500 with 100 routes per hostname",routes:500,routesPerHostname:100,phase:"scaling-up",throughput:5351.64,totalRequests:160555,latency:{min:370,mean:6797,max:91549,pstdev:12276,percentiles:{p50:3145,p75:5052,p80:5728,p90:8916,p95:47915,p99:54947,p999:65368}},resources:{envoyGateway:{memory:{min:186.4,max:200.27,mean:195.8},cpu:{min:.4,max:26.07,mean:1.71}},envoyProxy:{memory:{min:71.16,max:77.36,mean:77},cpu:{min:0,max:94.88,mean:3.39}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling up httproutes to 1000 with 200 routes per hostname",routes:1e3,routesPerHostname:200,phase:"scaling-up",throughput:5239.17,totalRequests:157179,latency:{min:367,mean:6908,max:97939,pstdev:12608,percentiles:{p50:3176,p75:5046,p80:5680,p90:8770,p95:48472,p99:56449,p999:68317}},resources:{envoyGateway:{memory:{min:220.03,max:237.27,mean:233.38},cpu:{min:.07,max:1.2,mean:.78}},envoyProxy:{memory:{min:129.13,max:129.59,mean:129.34},cpu:{min:0,max:74.43,mean:2.24}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling down httproutes to 10 with 2 routes per hostname",routes:10,routesPerHostname:2,phase:"scaling-down",throughput:5409.79,totalRequests:162296,latency:{min:386,mean:6690,max:71688,pstdev:12144,percentiles:{p50:3075,p75:4937,p80:5610,p90:8727,p95:48252,p99:54595,p999:59408}},resources:{envoyGateway:{memory:{min:165.91,max:176.04,mean:168.11},cpu:{min:.67,max:3.73,mean:1.23}},envoyProxy:{memory:{min:123.28,max:123.45,mean:123.35},cpu:{min:0,max:99.72,mean:9.08}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling down httproutes to 50 with 10 routes per hostname",routes:50,routesPerHostname:10,phase:"scaling-down",throughput:5480.62,totalRequests:164426,latency:{min:378,mean:7119,max:85032,pstdev:12404,percentiles:{p50:3390,p75:5340,p80:6015,p90:9300,p95:48457,p99:55005,p999:60284}},resources:{envoyGateway:{memory:{min:165.36,max:171.88,mean:169.47},cpu:{min:.6,max:6.8,mean:1.44}},envoyProxy:{memory:{min:123.27,max:123.97,mean:123.42},cpu:{min:0,max:100.05,mean:11.5}}},poolOverflow:359,upstreamConnections:41},{testName:"scaling down httproutes to 100 with 20 routes per hostname",routes:100,routesPerHostname:20,phase:"scaling-down",throughput:5413.63,totalRequests:162409,latency:{min:357,mean:6664,max:77307,pstdev:12030,percentiles:{p50:3140,p75:4935,p80:5560,p90:8456,p95:47661,p99:54544,p999:61110}},resources:{envoyGateway:{memory:{min:170.86,max:186.68,mean:173.53},cpu:{min:.67,max:67.27,mean:7.29}},envoyProxy:{memory:{min:123.27,max:129.61,mean:124.04},cpu:{min:0,max:100,mean:5.99}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling down httproutes to 300 with 60 routes per hostname",routes:300,routesPerHostname:60,phase:"scaling-down",throughput:5356.13,totalRequests:160684,latency:{min:376,mean:6544,max:86540,pstdev:11908,percentiles:{p50:3079,p75:4927,p80:5554,p90:8360,p95:47069,p99:54265,p999:62601}},resources:{envoyGateway:{memory:{min:176.87,max:192.99,mean:189.97},cpu:{min:.6,max:73.6,mean:5.05}},envoyProxy:{memory:{min:129.54,max:129.76,mean:129.65},cpu:{min:0,max:100.08,mean:11.45}}},poolOverflow:363,upstreamConnections:37},{testName:"scaling down httproutes to 500 with 100 routes per hostname",routes:500,routesPerHostname:100,phase:"scaling-down",throughput:5339.99,totalRequests:160200,latency:{min:395,mean:6772,max:99700,pstdev:12345,percentiles:{p50:3118,p75:4954,p80:5598,p90:8672,p95:48261,p99:55498,p999:65902}},resources:{envoyGateway:{memory:{min:199.81,max:210.76,mean:205.85},cpu:{min:.4,max:31.13,mean:2.01}},envoyProxy:{memory:{min:129.54,max:129.74,mean:129.6},cpu:{min:0,max:99.87,mean:8.99}}},poolOverflow:362,upstreamConnections:38}]},foe={metadata:{version:"1.4.1",runId:"1.4.1-1750189826783",date:"2025-06-04",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.4.1",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.4.1/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scaling up httproutes to 10 with 2 routes per hostname",routes:10,routesPerHostname:2,phase:"scaling-up",throughput:5440.31,totalRequests:163209,latency:{min:335,mean:6565,max:66668,pstdev:11480,percentiles:{p50:3258,p75:5079,p80:5722,p90:8679,p95:45924,p99:53512,p999:58220}},resources:{envoyGateway:{memory:{min:128.02,max:151.26,mean:147.41},cpu:{min:.27,max:.67,mean:.45}},envoyProxy:{memory:{min:0,max:26.92,mean:22.58},cpu:{min:0,max:99.73,mean:6.02}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling up httproutes to 50 with 10 routes per hostname",routes:50,routesPerHostname:10,phase:"scaling-up",throughput:5429.88,totalRequests:162900,latency:{min:345,mean:6468,max:70791,pstdev:11793,percentiles:{p50:3170,p75:4794,p80:5311,p90:7534,p95:47400,p99:53536,p999:58034}},resources:{envoyGateway:{memory:{min:151.26,max:170.78,mean:161.92},cpu:{min:.4,max:4.33,mean:.97}},envoyProxy:{memory:{min:26.73,max:33.13,mean:31.63},cpu:{min:0,max:99.94,mean:3.35}}},poolOverflow:363,upstreamConnections:37},{testName:"scaling up httproutes to 100 with 20 routes per hostname",routes:100,routesPerHostname:20,phase:"scaling-up",throughput:5499.86,totalRequests:164996,latency:{min:391,mean:8147,max:99663,pstdev:13319,percentiles:{p50:3957,p75:6308,p80:7120,p90:11799,p95:49942,p99:56190,p999:62709}},resources:{envoyGateway:{memory:{min:161.19,max:165.52,mean:163.96},cpu:{min:.4,max:8.67,mean:1.36}},envoyProxy:{memory:{min:32.98,max:39.18,mean:38.14},cpu:{min:0,max:99.97,mean:8.67}}},poolOverflow:353,upstreamConnections:47},{testName:"scaling up httproutes to 300 with 60 routes per hostname",routes:300,routesPerHostname:60,phase:"scaling-up",throughput:5335.1,totalRequests:160053,latency:{min:365,mean:6734,max:90963,pstdev:12096,percentiles:{p50:3228,p75:5042,p80:5649,p90:8465,p95:47704,p99:54396,p999:61714}},resources:{envoyGateway:{memory:{min:183.87,max:223.25,mean:206.96},cpu:{min:.47,max:79.87,mean:3.9}},envoyProxy:{memory:{min:57.15,max:57.52,mean:57.34},cpu:{min:0,max:99.97,mean:13.49}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling up httproutes to 500 with 100 routes per hostname",routes:500,routesPerHostname:100,phase:"scaling-up",throughput:5256.52,totalRequests:157696,latency:{min:389,mean:6320,max:77848,pstdev:11829,percentiles:{p50:2974,p75:4705,p80:5292,p90:7729,p95:47173,p99:54317,p999:64679}},resources:{envoyGateway:{memory:{min:186.16,max:199.19,mean:195.68},cpu:{min:.4,max:1.2,mean:.75}},envoyProxy:{memory:{min:63.26,max:79.43,mean:79.04},cpu:{min:0,max:99.86,mean:17.64}}},poolOverflow:365,upstreamConnections:35},{testName:"scaling up httproutes to 1000 with 200 routes per hostname",routes:1e3,routesPerHostname:200,phase:"scaling-up",throughput:5280.09,totalRequests:158409,latency:{min:387,mean:6871,max:94064,pstdev:12441,percentiles:{p50:3186,p75:5019,p80:5651,p90:8766,p95:48361,p99:55758,p999:67403}},resources:{envoyGateway:{memory:{min:230.81,max:243.68,mean:238.65},cpu:{min:.13,max:1.13,mean:.8}},envoyProxy:{memory:{min:127.63,max:127.86,mean:127.72},cpu:{min:0,max:85.57,mean:6.19}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling down httproutes to 10 with 2 routes per hostname",routes:10,routesPerHostname:2,phase:"scaling-down",throughput:5430.18,totalRequests:162909,latency:{min:376,mean:6679,max:79630,pstdev:12142,percentiles:{p50:3147,p75:5015,p80:5630,p90:8374,p95:48377,p99:54654,p999:59858}},resources:{envoyGateway:{memory:{min:164.61,max:169.75,mean:165.39},cpu:{min:.8,max:3.47,mean:1.14}},envoyProxy:{memory:{min:121.22,max:121.61,mean:121.3},cpu:{min:0,max:99.74,mean:8.02}}},poolOverflow:362,upstreamConnections:38},{testName:"scaling down httproutes to 50 with 10 routes per hostname",routes:50,routesPerHostname:10,phase:"scaling-down",throughput:5506.49,totalRequests:165195,latency:{min:386,mean:8319,max:85860,pstdev:13567,percentiles:{p50:3951,p75:6305,p80:7155,p90:12741,p95:50714,p99:57303,p999:63928}},resources:{envoyGateway:{memory:{min:165.63,max:175.23,mean:169.06},cpu:{min:.8,max:7.53,mean:1.54}},envoyProxy:{memory:{min:111.91,max:121.36,mean:113.46},cpu:{min:0,max:99.75,mean:7.62}}},poolOverflow:352,upstreamConnections:48},{testName:"scaling down httproutes to 100 with 20 routes per hostname",routes:100,routesPerHostname:20,phase:"scaling-down",throughput:5452.96,totalRequests:163589,latency:{min:395,mean:7694,max:97460,pstdev:12949,percentiles:{p50:3636,p75:5828,p80:6627,p90:10929,p95:49506,p99:55681,p999:62369}},resources:{envoyGateway:{memory:{min:166.64,max:195.48,mean:174.79},cpu:{min:.8,max:38.2,mean:5.63}},envoyProxy:{memory:{min:111.8,max:112.36,mean:111.95},cpu:{min:0,max:35.17,mean:2.77}}},poolOverflow:356,upstreamConnections:44},{testName:"scaling down httproutes to 300 with 60 routes per hostname",routes:300,routesPerHostname:60,phase:"scaling-down",throughput:5360.33,totalRequests:160815,latency:{min:354,mean:6538,max:92344,pstdev:12045,percentiles:{p50:3068,p75:4770,p80:5341,p90:7974,p95:47742,p99:54480,p999:62144}},resources:{envoyGateway:{memory:{min:180.27,max:192.65,mean:189.01},cpu:{min:.67,max:46.53,mean:4.41}},envoyProxy:{memory:{min:127.66,max:127.84,mean:127.72},cpu:{min:0,max:100.03,mean:9.14}}},poolOverflow:363,upstreamConnections:37},{testName:"scaling down httproutes to 500 with 100 routes per hostname",routes:500,routesPerHostname:100,phase:"scaling-down",throughput:5334.03,totalRequests:160024,latency:{min:392,mean:6782,max:132493,pstdev:12373,percentiles:{p50:3182,p75:4977,p80:5586,p90:8359,p95:48365,p99:55252,p999:64712}},resources:{envoyGateway:{memory:{min:203.37,max:217.03,mean:209.04},cpu:{min:.67,max:5.67,mean:1.33}},envoyProxy:{memory:{min:127.65,max:127.84,mean:127.72},cpu:{min:0,max:99.86,mean:9.91}}},poolOverflow:362,upstreamConnections:38}]},doe={metadata:{version:"1.3.2",runId:"1.3.2-1750189826777",date:"2025-03-24",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.3.2",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.3.2/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5203.06,totalRequests:156092,latency:{min:371,mean:6970,max:73949,pstdev:12106,percentiles:{p50:3331,p75:5396,p80:6080,p90:9715,p95:47564,p99:55666,p999:60872}},resources:{envoyGateway:{memory:{min:126.68,max:126.68,mean:126.68},cpu:{min:.99,max:.99,mean:.99}},envoyProxy:{memory:{min:28.17,max:28.17,mean:28.17},cpu:{min:30.49,max:30.49,mean:30.49}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5157.35,totalRequests:154724,latency:{min:380,mean:6752,max:83992,pstdev:12526,percentiles:{p50:3089,p75:4875,p80:5484,p90:8253,p95:49682,p99:55398,p999:61507}},resources:{envoyGateway:{memory:{min:129.05,max:129.05,mean:129.05},cpu:{min:1.79,max:1.79,mean:1.79}},envoyProxy:{memory:{min:32.33,max:32.33,mean:32.33},cpu:{min:61.02,max:61.02,mean:61.02}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5052.19,totalRequests:151566,latency:{min:363,mean:7564,max:72122,pstdev:13492,percentiles:{p50:3384,p75:5472,p80:6191,p90:9941,p95:51320,p99:57128,p999:63107}},resources:{envoyGateway:{memory:{min:146.92,max:146.92,mean:146.92},cpu:{min:3.17,max:3.17,mean:3.17}},envoyProxy:{memory:{min:40.52,max:40.52,mean:40.52},cpu:{min:91.79,max:91.79,mean:91.79}}},poolOverflow:359,upstreamConnections:41},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5101.42,totalRequests:153051,latency:{min:397,mean:6832,max:82624,pstdev:12894,percentiles:{p50:2990,p75:4798,p80:5418,p90:8366,p95:50391,p99:56737,p999:65114}},resources:{envoyGateway:{memory:{min:154.98,max:154.98,mean:154.98},cpu:{min:15.2,max:15.2,mean:15.2}},envoyProxy:{memory:{min:58.7,max:58.7,mean:58.7},cpu:{min:124.54,max:124.54,mean:124.54}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:4948.44,totalRequests:148459,latency:{min:398,mean:7200,max:94949,pstdev:13256,percentiles:{p50:3208,p75:5057,p80:5652,p90:8780,p95:50821,p99:57442,p999:66746}},resources:{envoyGateway:{memory:{min:163.01,max:163.01,mean:163.01},cpu:{min:27.68,max:27.68,mean:27.68}},envoyProxy:{memory:{min:78.9,max:78.9,mean:78.9},cpu:{min:158.2,max:158.2,mean:158.2}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:4933.76,totalRequests:148015,latency:{min:361,mean:7088,max:127959,pstdev:13349,percentiles:{p50:3052,p75:4885,p80:5529,p90:8785,p95:50700,p99:59123,p999:70660}},resources:{envoyGateway:{memory:{min:196.05,max:196.05,mean:196.05},cpu:{min:62.32,max:62.32,mean:62.32}},envoyProxy:{memory:{min:127.19,max:127.19,mean:127.19},cpu:{min:196.79,max:196.79,mean:196.79}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5116.93,totalRequests:153515,latency:{min:387,mean:9017,max:78245,pstdev:15046,percentiles:{p50:3922,p75:6583,p80:7609,p90:16301,p95:54007,p99:59990,p999:67956}},resources:{envoyGateway:{memory:{min:210.06,max:210.06,mean:210.06},cpu:{min:115.56,max:115.56,mean:115.56}},envoyProxy:{memory:{min:110.31,max:110.31,mean:110.31},cpu:{min:358.89,max:358.89,mean:358.89}}},poolOverflow:351,upstreamConnections:49},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:4991.13,totalRequests:149734,latency:{min:391,mean:7071,max:80433,pstdev:12841,percentiles:{p50:3199,p75:5005,p80:5635,p90:8730,p95:49813,p99:55756,p999:61407}},resources:{envoyGateway:{memory:{min:199.41,max:199.41,mean:199.41},cpu:{min:114.89,max:114.89,mean:114.89}},envoyProxy:{memory:{min:127.44,max:127.44,mean:127.44},cpu:{min:328.33,max:328.33,mean:328.33}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5138.59,totalRequests:154161,latency:{min:377,mean:6954,max:86982,pstdev:13049,percentiles:{p50:3045,p75:4886,p80:5525,p90:8466,p95:51048,p99:56733,p999:62337}},resources:{envoyGateway:{memory:{min:219.76,max:219.76,mean:219.76},cpu:{min:113.71,max:113.71,mean:113.71}},envoyProxy:{memory:{min:127.31,max:127.31,mean:127.31},cpu:{min:297.73,max:297.73,mean:297.73}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5051.68,totalRequests:151551,latency:{min:389,mean:7977,max:86577,pstdev:14090,percentiles:{p50:3498,p75:5731,p80:6471,p90:10947,p95:52355,p99:58748,p999:67379}},resources:{envoyGateway:{memory:{min:259.14,max:259.14,mean:259.14},cpu:{min:103.22,max:103.22,mean:103.22}},envoyProxy:{memory:{min:127.31,max:127.31,mean:127.31},cpu:{min:265.59,max:265.59,mean:265.59}}},poolOverflow:357,upstreamConnections:43},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:4962.4,totalRequests:148872,latency:{min:384,mean:7209,max:110186,pstdev:13357,percentiles:{p50:3112,p75:5018,p80:5694,p90:9338,p95:51267,p99:57384,p999:67244}},resources:{envoyGateway:{memory:{min:177.76,max:177.76,mean:177.76},cpu:{min:91.86,max:91.86,mean:91.86}},envoyProxy:{memory:{min:127.24,max:127.24,mean:127.24},cpu:{min:233.07,max:233.07,mean:233.07}}},poolOverflow:362,upstreamConnections:38}]},moe={metadata:{version:"1.3.1",runId:"1.3.1-1750189826775",date:"2025-03-05",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.3.1",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.3.1/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5115.35,totalRequests:153465,latency:{min:389,mean:6604,max:74543,pstdev:11843,percentiles:{p50:3094,p75:5007,p80:5671,p90:8723,p95:46997,p99:55033,p999:60336}},resources:{envoyGateway:{memory:{min:122.23,max:122.23,mean:122.23},cpu:{min:1.04,max:1.04,mean:1.04}},envoyProxy:{memory:{min:25.28,max:25.28,mean:25.28},cpu:{min:30.51,max:30.51,mean:30.51}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5150.43,totalRequests:154513,latency:{min:386,mean:6919,max:87556,pstdev:12645,percentiles:{p50:3126,p75:5009,p80:5646,p90:8620,p95:49580,p99:55326,p999:60073}},resources:{envoyGateway:{memory:{min:151.59,max:151.59,mean:151.59},cpu:{min:1.83,max:1.83,mean:1.83}},envoyProxy:{memory:{min:31.45,max:31.45,mean:31.45},cpu:{min:61.11,max:61.11,mean:61.11}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5211.7,totalRequests:156354,latency:{min:376,mean:6687,max:88604,pstdev:12657,percentiles:{p50:2913,p75:4676,p80:5309,p90:8259,p95:49930,p99:55853,p999:63514}},resources:{envoyGateway:{memory:{min:145.61,max:145.61,mean:145.61},cpu:{min:3.19,max:3.19,mean:3.19}},envoyProxy:{memory:{min:35.5,max:35.5,mean:35.5},cpu:{min:91.72,max:91.72,mean:91.72}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5066.99,totalRequests:152016,latency:{min:386,mean:7628,max:93048,pstdev:13700,percentiles:{p50:3310,p75:5414,p80:6156,p90:10289,p95:51750,p99:58007,p999:66938}},resources:{envoyGateway:{memory:{min:156.12,max:156.12,mean:156.12},cpu:{min:15.73,max:15.73,mean:15.73}},envoyProxy:{memory:{min:55.56,max:55.56,mean:55.56},cpu:{min:124.54,max:124.54,mean:124.54}}},poolOverflow:359,upstreamConnections:41},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5098.53,totalRequests:152960,latency:{min:373,mean:7203,max:108142,pstdev:13209,percentiles:{p50:3161,p75:5083,p80:5791,p90:9292,p95:50372,p99:56864,p999:67850}},resources:{envoyGateway:{memory:{min:181.23,max:181.23,mean:181.23},cpu:{min:28.35,max:28.35,mean:28.35}},envoyProxy:{memory:{min:73.6,max:73.6,mean:73.6},cpu:{min:158.01,max:158.01,mean:158.01}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5048.96,totalRequests:151469,latency:{min:366,mean:7462,max:99799,pstdev:13676,percentiles:{p50:3122,p75:5245,p80:6050,p90:10398,p95:50685,p99:59639,p999:74547}},resources:{envoyGateway:{memory:{min:196.57,max:196.57,mean:196.57},cpu:{min:61.62,max:61.62,mean:61.62}},envoyProxy:{memory:{min:119.73,max:119.73,mean:119.73},cpu:{min:196.43,max:196.43,mean:196.43}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5116.49,totalRequests:153495,latency:{min:373,mean:7373,max:79749,pstdev:13231,percentiles:{p50:3334,p75:5268,p80:5935,p90:9436,p95:50972,p99:56680,p999:62969}},resources:{envoyGateway:{memory:{min:213.36,max:213.36,mean:213.36},cpu:{min:114.56,max:114.56,mean:114.56}},envoyProxy:{memory:{min:109.79,max:109.79,mean:109.79},cpu:{min:357.84,max:357.84,mean:357.84}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5128.63,totalRequests:153859,latency:{min:380,mean:6787,max:76259,pstdev:12813,percentiles:{p50:2982,p75:4725,p80:5330,p90:8194,p95:50354,p99:56367,p999:61741}},resources:{envoyGateway:{memory:{min:217.34,max:217.34,mean:217.34},cpu:{min:113.9,max:113.9,mean:113.9}},envoyProxy:{memory:{min:119.73,max:119.73,mean:119.73},cpu:{min:327.37,max:327.37,mean:327.37}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5217.06,totalRequests:156516,latency:{min:386,mean:7182,max:93491,pstdev:12999,percentiles:{p50:3211,p75:5188,p80:5872,p90:9449,p95:50083,p99:56135,p999:63309}},resources:{envoyGateway:{memory:{min:220.2,max:220.2,mean:220.2},cpu:{min:112.64,max:112.64,mean:112.64}},envoyProxy:{memory:{min:119.73,max:119.73,mean:119.73},cpu:{min:296.73,max:296.73,mean:296.73}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5033.4,totalRequests:151002,latency:{min:380,mean:6522,max:92626,pstdev:12472,percentiles:{p50:2849,p75:4643,p80:5272,p90:8064,p95:49350,p99:55799,p999:64323}},resources:{envoyGateway:{memory:{min:201.77,max:201.77,mean:201.77},cpu:{min:102.25,max:102.25,mean:102.25}},envoyProxy:{memory:{min:119.73,max:119.73,mean:119.73},cpu:{min:264.74,max:264.74,mean:264.74}}},poolOverflow:365,upstreamConnections:35},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5061.34,totalRequests:151843,latency:{min:393,mean:7429,max:99934,pstdev:13552,percentiles:{p50:3227,p75:5180,p80:5867,p90:9747,p95:51320,p99:58009,p999:68829}},resources:{envoyGateway:{memory:{min:173.04,max:173.04,mean:173.04},cpu:{min:91.12,max:91.12,mean:91.12}},envoyProxy:{memory:{min:119.73,max:119.73,mean:119.73},cpu:{min:231.98,max:231.98,mean:231.98}}},poolOverflow:360,upstreamConnections:40}]},hoe={metadata:{version:"1.3.0",runId:"1.3.0-1750189826772",date:"2025-01-31",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.3.0",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.3.0/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5297.23,totalRequests:158917,latency:{min:371,mean:6861,max:69562,pstdev:12096,percentiles:{p50:3223,p75:5219,p80:5926,p90:9360,p95:48029,p99:54945,p999:60096}},resources:{envoyGateway:{memory:{min:116.36,max:116.36,mean:116.36},cpu:{min:.77,max:.77,mean:.77}},envoyProxy:{memory:{min:25.3,max:25.3,mean:25.3},cpu:{min:30.42,max:30.42,mean:30.42}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5276.06,totalRequests:158285,latency:{min:372,mean:6781,max:81436,pstdev:12685,percentiles:{p50:3035,p75:4783,p80:5410,p90:8210,p95:50042,p99:55861,p999:61739}},resources:{envoyGateway:{memory:{min:135.97,max:135.97,mean:135.97},cpu:{min:1.53,max:1.53,mean:1.53}},envoyProxy:{memory:{min:31.45,max:31.45,mean:31.45},cpu:{min:60.97,max:60.97,mean:60.97}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5248.15,totalRequests:157445,latency:{min:386,mean:6832,max:83939,pstdev:12879,percentiles:{p50:2957,p75:4826,p80:5504,p90:8587,p95:50454,p99:56524,p999:63150}},resources:{envoyGateway:{memory:{min:128.67,max:128.67,mean:128.67},cpu:{min:2.88,max:2.88,mean:2.88}},envoyProxy:{memory:{min:35.62,max:35.62,mean:35.62},cpu:{min:91.69,max:91.69,mean:91.69}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5190.05,totalRequests:155708,latency:{min:374,mean:6735,max:105820,pstdev:12667,percentiles:{p50:2997,p75:4715,p80:5323,p90:8205,p95:49692,p99:56199,p999:64403}},resources:{envoyGateway:{memory:{min:147.27,max:147.27,mean:147.27},cpu:{min:15.8,max:15.8,mean:15.8}},envoyProxy:{memory:{min:61.52,max:61.52,mean:61.52},cpu:{min:124.52,max:124.52,mean:124.52}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5171.63,totalRequests:155149,latency:{min:345,mean:7139,max:89923,pstdev:13210,percentiles:{p50:3191,p75:5039,p80:5680,p90:8824,p95:50767,p99:57397,p999:67665}},resources:{envoyGateway:{memory:{min:158.88,max:158.88,mean:158.88},cpu:{min:28.37,max:28.37,mean:28.37}},envoyProxy:{memory:{min:76,max:76,mean:76},cpu:{min:157.59,max:157.59,mean:157.59}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:4998.06,totalRequests:149944,latency:{min:371,mean:6950,max:98525,pstdev:13111,percentiles:{p50:3057,p75:4816,p80:5416,p90:8383,p95:50270,p99:57946,p999:69038}},resources:{envoyGateway:{memory:{min:182.71,max:182.71,mean:182.71},cpu:{min:61.79,max:61.79,mean:61.79}},envoyProxy:{memory:{min:122.12,max:122.12,mean:122.12},cpu:{min:194.7,max:194.7,mean:194.7}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5224.35,totalRequests:156731,latency:{min:376,mean:6870,max:73203,pstdev:12827,percentiles:{p50:3060,p75:4817,p80:5428,p90:8280,p95:50618,p99:56178,p999:61280}},resources:{envoyGateway:{memory:{min:209.85,max:209.85,mean:209.85},cpu:{min:114.89,max:114.89,mean:114.89}},envoyProxy:{memory:{min:110.5,max:110.5,mean:110.5},cpu:{min:356.13,max:356.13,mean:356.13}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5247.75,totalRequests:157436,latency:{min:383,mean:6991,max:88027,pstdev:12998,percentiles:{p50:3091,p75:4901,p80:5532,p90:8568,p95:50655,p99:56272,p999:62160}},resources:{envoyGateway:{memory:{min:208.95,max:208.95,mean:208.95},cpu:{min:114.35,max:114.35,mean:114.35}},envoyProxy:{memory:{min:122.19,max:122.19,mean:122.19},cpu:{min:325.78,max:325.78,mean:325.78}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5201.69,totalRequests:156051,latency:{min:380,mean:7027,max:78663,pstdev:12995,percentiles:{p50:3129,p75:4963,p80:5600,p90:8642,p95:50726,p99:56254,p999:61509}},resources:{envoyGateway:{memory:{min:211.34,max:211.34,mean:211.34},cpu:{min:113.23,max:113.23,mean:113.23}},envoyProxy:{memory:{min:122.17,max:122.17,mean:122.17},cpu:{min:295.15,max:295.15,mean:295.15}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5206.95,totalRequests:156214,latency:{min:386,mean:6893,max:92049,pstdev:12874,percentiles:{p50:3088,p75:4898,p80:5548,p90:8444,p95:50268,p99:56334,p999:65038}},resources:{envoyGateway:{memory:{min:158.73,max:158.73,mean:158.73},cpu:{min:102.79,max:102.79,mean:102.79}},envoyProxy:{memory:{min:122.15,max:122.15,mean:122.15},cpu:{min:263,max:263,mean:263}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5150.23,totalRequests:154507,latency:{min:375,mean:6786,max:84877,pstdev:12812,percentiles:{p50:3009,p75:4734,p80:5340,p90:8236,p95:50247,p99:56647,p999:65042}},resources:{envoyGateway:{memory:{min:168.04,max:168.04,mean:168.04},cpu:{min:91.51,max:91.51,mean:91.51}},envoyProxy:{memory:{min:122.14,max:122.14,mean:122.14},cpu:{min:230.49,max:230.49,mean:230.49}}},poolOverflow:363,upstreamConnections:37}]},voe={metadata:{version:"1.2.8",runId:"1.2.8-1750189826771",date:"2025-03-25",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.8",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.8/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5878.23,totalRequests:176347,latency:{min:376,mean:9154,max:75661,pstdev:14181,percentiles:{p50:4294,p75:7182,p80:8292,p90:22716,p95:52056,p99:58071,p999:64841}},resources:{envoyGateway:{memory:{min:108.89,max:108.89,mean:108.89},cpu:{min:.77,max:.77,mean:.77}},envoyProxy:{memory:{min:28.23,max:28.23,mean:28.23},cpu:{min:30.48,max:30.48,mean:30.48}}},poolOverflow:342,upstreamConnections:58},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5582.21,totalRequests:167469,latency:{min:359,mean:6378,max:69967,pstdev:12106,percentiles:{p50:2864,p75:4529,p80:5113,p90:7717,p95:48431,p99:54521,p999:59967}},resources:{envoyGateway:{memory:{min:132.62,max:132.62,mean:132.62},cpu:{min:1.54,max:1.54,mean:1.54}},envoyProxy:{memory:{min:32.39,max:32.39,mean:32.39},cpu:{min:60.99,max:60.99,mean:60.99}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5454.1,totalRequests:163623,latency:{min:368,mean:6738,max:72437,pstdev:12216,percentiles:{p50:3318,p75:5037,p80:5581,p90:8092,p95:47886,p99:53829,p999:60200}},resources:{envoyGateway:{memory:{min:135.66,max:135.66,mean:135.66},cpu:{min:2.91,max:2.91,mean:2.91}},envoyProxy:{memory:{min:38.6,max:38.6,mean:38.6},cpu:{min:91.91,max:91.91,mean:91.91}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5496.5,totalRequests:164898,latency:{min:375,mean:6453,max:103043,pstdev:12191,percentiles:{p50:2887,p75:4603,p80:5209,p90:7967,p95:48150,p99:54630,p999:63743}},resources:{envoyGateway:{memory:{min:147.15,max:147.15,mean:147.15},cpu:{min:15.47,max:15.47,mean:15.47}},envoyProxy:{memory:{min:60.75,max:60.75,mean:60.75},cpu:{min:125.31,max:125.31,mean:125.31}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5351.8,totalRequests:160554,latency:{min:377,mean:6516,max:93298,pstdev:12549,percentiles:{p50:2857,p75:4496,p80:5076,p90:7615,p95:49512,p99:55822,p999:65318}},resources:{envoyGateway:{memory:{min:155.79,max:155.79,mean:155.79},cpu:{min:28.01,max:28.01,mean:28.01}},envoyProxy:{memory:{min:82.96,max:82.96,mean:82.96},cpu:{min:159.29,max:159.29,mean:159.29}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5102.69,totalRequests:153086,latency:{min:358,mean:6968,max:108298,pstdev:13304,percentiles:{p50:2979,p75:4802,p80:5441,p90:8343,p95:50370,p99:60203,p999:71446}},resources:{envoyGateway:{memory:{min:190.03,max:190.03,mean:190.03},cpu:{min:61.69,max:61.69,mean:61.69}},envoyProxy:{memory:{min:133.22,max:133.22,mean:133.22},cpu:{min:199.84,max:199.84,mean:199.84}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5424.03,totalRequests:162721,latency:{min:362,mean:6570,max:73609,pstdev:12562,percentiles:{p50:2923,p75:4626,p80:5209,p90:7751,p95:49960,p99:55525,p999:60127}},resources:{envoyGateway:{memory:{min:127.89,max:127.89,mean:127.89},cpu:{min:114.69,max:114.69,mean:114.69}},envoyProxy:{memory:{min:112.11,max:112.11,mean:112.11},cpu:{min:363.01,max:363.01,mean:363.01}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5454.96,totalRequests:163653,latency:{min:363,mean:6397,max:73404,pstdev:12347,percentiles:{p50:2884,p75:4452,p80:4989,p90:7300,p95:49477,p99:55369,p999:61495}},resources:{envoyGateway:{memory:{min:136.28,max:136.28,mean:136.28},cpu:{min:114.05,max:114.05,mean:114.05}},envoyProxy:{memory:{min:114.79,max:114.79,mean:114.79},cpu:{min:332.54,max:332.54,mean:332.54}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5402.5,totalRequests:162075,latency:{min:366,mean:6575,max:127266,pstdev:12484,percentiles:{p50:2940,p75:4600,p80:5162,p90:7727,p95:49549,p99:55334,p999:61478}},resources:{envoyGateway:{memory:{min:140.91,max:140.91,mean:140.91},cpu:{min:112.87,max:112.87,mean:112.87}},envoyProxy:{memory:{min:114.76,max:114.76,mean:114.76},cpu:{min:301.85,max:301.85,mean:301.85}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5498.81,totalRequests:164969,latency:{min:349,mean:6325,max:84213,pstdev:12296,percentiles:{p50:2789,p75:4396,p80:4944,p90:7293,p95:48961,p99:55212,p999:65019}},resources:{envoyGateway:{memory:{min:148.36,max:148.36,mean:148.36},cpu:{min:102.53,max:102.53,mean:102.53}},envoyProxy:{memory:{min:133.35,max:133.35,mean:133.35},cpu:{min:269.43,max:269.43,mean:269.43}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5308.76,totalRequests:159263,latency:{min:361,mean:6727,max:98807,pstdev:12828,percentiles:{p50:2916,p75:4650,p80:5273,p90:8161,p95:49971,p99:56840,p999:67182}},resources:{envoyGateway:{memory:{min:157.93,max:157.93,mean:157.93},cpu:{min:91.26,max:91.26,mean:91.26}},envoyProxy:{memory:{min:133.2,max:133.2,mean:133.2},cpu:{min:236.02,max:236.02,mean:236.02}}},poolOverflow:362,upstreamConnections:38}]},yoe={metadata:{version:"1.2.7",runId:"1.2.7-1750189826769",date:"2025-03-06",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.7",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.7/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5984.83,totalRequests:179545,latency:{min:402,mean:12033,max:92835,pstdev:16725,percentiles:{p50:5655,p75:9809,p80:11804,p90:49971,p95:55760,p99:64241,p999:73961}},resources:{envoyGateway:{memory:{min:112.35,max:112.35,mean:112.35},cpu:{min:.76,max:.76,mean:.76}},envoyProxy:{memory:{min:28.29,max:28.29,mean:28.29},cpu:{min:30.49,max:30.49,mean:30.49}}},poolOverflow:323,upstreamConnections:77},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5564.85,totalRequests:166949,latency:{min:362,mean:6416,max:73261,pstdev:12268,percentiles:{p50:2855,p75:4514,p80:5094,p90:7660,p95:49031,p99:54939,p999:60436}},resources:{envoyGateway:{memory:{min:125.94,max:125.94,mean:125.94},cpu:{min:1.52,max:1.52,mean:1.52}},envoyProxy:{memory:{min:34.46,max:34.46,mean:34.46},cpu:{min:61.11,max:61.11,mean:61.11}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5618.23,totalRequests:168547,latency:{min:372,mean:6834,max:85766,pstdev:12643,percentiles:{p50:3107,p75:4954,p80:5560,p90:8454,p95:49485,p99:55513,p999:61272}},resources:{envoyGateway:{memory:{min:132.05,max:132.05,mean:132.05},cpu:{min:2.92,max:2.92,mean:2.92}},envoyProxy:{memory:{min:40.63,max:40.63,mean:40.63},cpu:{min:92.02,max:92.02,mean:92.02}}},poolOverflow:359,upstreamConnections:41},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5506.26,totalRequests:165193,latency:{min:383,mean:6670,max:85905,pstdev:12677,percentiles:{p50:2946,p75:4659,p80:5243,p90:7891,p95:49846,p99:56096,p999:65185}},resources:{envoyGateway:{memory:{min:146.09,max:146.09,mean:146.09},cpu:{min:15.43,max:15.43,mean:15.43}},envoyProxy:{memory:{min:62.86,max:62.86,mean:62.86},cpu:{min:125.21,max:125.21,mean:125.21}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5580.92,totalRequests:167428,latency:{min:310,mean:6716,max:89178,pstdev:12233,percentiles:{p50:3157,p75:4861,p80:5446,p90:8231,p95:47419,p99:54634,p999:65587}},resources:{envoyGateway:{memory:{min:163.49,max:163.49,mean:163.49},cpu:{min:28.15,max:28.15,mean:28.15}},envoyProxy:{memory:{min:83.07,max:83.07,mean:83.07},cpu:{min:159.15,max:159.15,mean:159.15}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5298.3,totalRequests:158951,latency:{min:369,mean:6537,max:122015,pstdev:12654,percentiles:{p50:2914,p75:4472,p80:5012,p90:7402,p95:48639,p99:58097,p999:69103}},resources:{envoyGateway:{memory:{min:187.81,max:187.81,mean:187.81},cpu:{min:62.41,max:62.41,mean:62.41}},envoyProxy:{memory:{min:133.27,max:133.27,mean:133.27},cpu:{min:199.13,max:199.13,mean:199.13}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5623.57,totalRequests:168711,latency:{min:380,mean:7190,max:75063,pstdev:13103,percentiles:{p50:3211,p75:5102,p80:5772,p90:9144,p95:50722,p99:56260,p999:61515}},resources:{envoyGateway:{memory:{min:132.38,max:132.38,mean:132.38},cpu:{min:116.08,max:116.08,mean:116.08}},envoyProxy:{memory:{min:128.69,max:128.69,mean:128.69},cpu:{min:362.3,max:362.3,mean:362.3}}},poolOverflow:357,upstreamConnections:43},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5516.59,totalRequests:165498,latency:{min:329,mean:6044,max:69902,pstdev:11632,percentiles:{p50:2893,p75:4337,p80:4790,p90:6742,p95:47155,p99:53372,p999:58073}},resources:{envoyGateway:{memory:{min:144.02,max:144.02,mean:144.02},cpu:{min:115.5,max:115.5,mean:115.5}},envoyProxy:{memory:{min:129.14,max:129.14,mean:129.14},cpu:{min:331.76,max:331.76,mean:331.76}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5515.82,totalRequests:165475,latency:{min:380,mean:6483,max:79159,pstdev:12363,percentiles:{p50:2872,p75:4545,p80:5131,p90:7760,p95:49190,p99:55408,p999:61890}},resources:{envoyGateway:{memory:{min:135.79,max:135.79,mean:135.79},cpu:{min:114.36,max:114.36,mean:114.36}},envoyProxy:{memory:{min:109,max:109,mean:109},cpu:{min:301.13,max:301.13,mean:301.13}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5403.53,totalRequests:162106,latency:{min:354,mean:6429,max:83423,pstdev:12305,percentiles:{p50:2864,p75:4496,p80:5072,p90:7504,p95:48885,p99:55414,p999:63565}},resources:{envoyGateway:{memory:{min:155.42,max:155.42,mean:155.42},cpu:{min:103.9,max:103.9,mean:103.9}},envoyProxy:{memory:{min:133.59,max:133.59,mean:133.59},cpu:{min:268.85,max:268.85,mean:268.85}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5405.24,totalRequests:162162,latency:{min:387,mean:6615,max:90812,pstdev:12660,percentiles:{p50:2913,p75:4593,p80:5174,p90:7866,p95:49461,p99:56510,p999:67248}},resources:{envoyGateway:{memory:{min:165.58,max:165.58,mean:165.58},cpu:{min:92.35,max:92.35,mean:92.35}},envoyProxy:{memory:{min:133.42,max:133.42,mean:133.42},cpu:{min:235.87,max:235.87,mean:235.87}}},poolOverflow:362,upstreamConnections:38}]},goe={metadata:{version:"1.2.6",runId:"1.2.6-1750189826767",date:"2025-01-23",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.6",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.6/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5668.03,totalRequests:170041,latency:{min:354,mean:6713,max:73658,pstdev:11906,percentiles:{p50:3142,p75:5170,p80:5865,p90:9278,p95:47425,p99:54870,p999:59602}},resources:{envoyGateway:{memory:{min:108.82,max:108.82,mean:108.82},cpu:{min:.76,max:.76,mean:.76}},envoyProxy:{memory:{min:25.38,max:25.38,mean:25.38},cpu:{min:30.51,max:30.51,mean:30.51}}},poolOverflow:359,upstreamConnections:41},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5542.06,totalRequests:166262,latency:{min:380,mean:6117,max:73658,pstdev:11949,percentiles:{p50:2749,p75:4310,p80:4842,p90:7067,p95:48484,p99:54499,p999:59725}},resources:{envoyGateway:{memory:{min:135.08,max:135.08,mean:135.08},cpu:{min:1.53,max:1.53,mean:1.53}},envoyProxy:{memory:{min:31.53,max:31.53,mean:31.53},cpu:{min:61.17,max:61.17,mean:61.17}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5556.06,totalRequests:166682,latency:{min:363,mean:6439,max:100163,pstdev:12297,percentiles:{p50:2891,p75:4506,p80:5050,p90:7595,p95:49078,p99:54935,p999:60456}},resources:{envoyGateway:{memory:{min:140.09,max:140.09,mean:140.09},cpu:{min:2.88,max:2.88,mean:2.88}},envoyProxy:{memory:{min:37.7,max:37.7,mean:37.7},cpu:{min:91.98,max:91.98,mean:91.98}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5462.58,totalRequests:163882,latency:{min:372,mean:6376,max:78544,pstdev:12289,percentiles:{p50:2853,p75:4454,p80:5002,p90:7458,p95:49076,p99:55132,p999:63879}},resources:{envoyGateway:{memory:{min:146.71,max:146.71,mean:146.71},cpu:{min:15.64,max:15.64,mean:15.64}},envoyProxy:{memory:{min:59.89,max:59.89,mean:59.89},cpu:{min:125.41,max:125.41,mean:125.41}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5427.5,totalRequests:162825,latency:{min:359,mean:6562,max:97148,pstdev:12598,percentiles:{p50:2881,p75:4518,p80:5091,p90:7774,p95:49301,p99:56102,p999:67338}},resources:{envoyGateway:{memory:{min:155.26,max:155.26,mean:155.26},cpu:{min:28.52,max:28.52,mean:28.52}},envoyProxy:{memory:{min:82.08,max:82.08,mean:82.08},cpu:{min:159.72,max:159.72,mean:159.72}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5261.62,totalRequests:157849,latency:{min:373,mean:6972,max:115060,pstdev:13251,percentiles:{p50:3015,p75:4770,p80:5385,p90:8293,p95:50296,p99:59930,p999:70590}},resources:{envoyGateway:{memory:{min:188.13,max:188.13,mean:188.13},cpu:{min:61.78,max:61.78,mean:61.78}},envoyProxy:{memory:{min:130.35,max:130.35,mean:130.35},cpu:{min:199.1,max:199.1,mean:199.1}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5543.17,totalRequests:166299,latency:{min:360,mean:6447,max:75276,pstdev:12396,percentiles:{p50:2866,p75:4486,p80:5040,p90:7527,p95:49663,p99:55240,p999:60647}},resources:{envoyGateway:{memory:{min:129.36,max:129.36,mean:129.36},cpu:{min:115.21,max:115.21,mean:115.21}},envoyProxy:{memory:{min:122.58,max:122.58,mean:122.58},cpu:{min:362.6,max:362.6,mean:362.6}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5522.73,totalRequests:165682,latency:{min:384,mean:6098,max:68321,pstdev:11922,percentiles:{p50:2776,p75:4326,p80:4838,p90:6958,p95:48431,p99:54296,p999:58857}},resources:{envoyGateway:{memory:{min:140.97,max:140.97,mean:140.97},cpu:{min:114.57,max:114.57,mean:114.57}},envoyProxy:{memory:{min:122.83,max:122.83,mean:122.83},cpu:{min:332.12,max:332.12,mean:332.12}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5500.66,totalRequests:165023,latency:{min:360,mean:6301,max:80400,pstdev:12064,percentiles:{p50:2833,p75:4483,p80:5035,p90:7574,p95:48777,p99:54876,p999:60850}},resources:{envoyGateway:{memory:{min:132.45,max:132.45,mean:132.45},cpu:{min:113.5,max:113.5,mean:113.5}},envoyProxy:{memory:{min:122.84,max:122.84,mean:122.84},cpu:{min:301.26,max:301.26,mean:301.26}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5453.33,totalRequests:163603,latency:{min:357,mean:6324,max:86024,pstdev:12138,percentiles:{p50:2842,p75:4429,p80:4967,p90:7393,p95:48201,p99:54683,p999:65251}},resources:{envoyGateway:{memory:{min:150.62,max:150.62,mean:150.62},cpu:{min:103.11,max:103.11,mean:103.11}},envoyProxy:{memory:{min:130.34,max:130.34,mean:130.34},cpu:{min:269.03,max:269.03,mean:269.03}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5402.96,totalRequests:162089,latency:{min:371,mean:6575,max:90804,pstdev:12621,percentiles:{p50:2877,p75:4543,p80:5114,p90:7832,p95:49375,p99:56092,p999:69132}},resources:{envoyGateway:{memory:{min:159.25,max:159.25,mean:159.25},cpu:{min:91.66,max:91.66,mean:91.66}},envoyProxy:{memory:{min:130.34,max:130.34,mean:130.34},cpu:{min:235.83,max:235.83,mean:235.83}}},poolOverflow:362,upstreamConnections:38}]},xoe={metadata:{version:"1.2.5",runId:"1.2.5-1750189826766",date:"2025-01-14",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.5",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.5/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5704.46,totalRequests:171136,latency:{min:370,mean:6520,max:74477,pstdev:11781,percentiles:{p50:3051,p75:4913,p80:5565,p90:8700,p95:47091,p99:54638,p999:59404}},resources:{envoyGateway:{memory:{min:112.07,max:112.07,mean:112.07},cpu:{min:.76,max:.76,mean:.76}},envoyProxy:{memory:{min:24.38,max:24.38,mean:24.38},cpu:{min:30.47,max:30.47,mean:30.47}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5647.86,totalRequests:169439,latency:{min:379,mean:6313,max:80498,pstdev:11935,percentiles:{p50:2854,p75:4550,p80:5158,p90:7820,p95:48025,p99:54390,p999:60321}},resources:{envoyGateway:{memory:{min:123.4,max:123.4,mean:123.4},cpu:{min:1.48,max:1.48,mean:1.48}},envoyProxy:{memory:{min:30.55,max:30.55,mean:30.55},cpu:{min:61.09,max:61.09,mean:61.09}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5529.29,totalRequests:165879,latency:{min:379,mean:6240,max:73539,pstdev:12017,percentiles:{p50:2818,p75:4510,p80:5068,p90:7413,p95:48300,p99:54171,p999:59308}},resources:{envoyGateway:{memory:{min:139.7,max:139.7,mean:139.7},cpu:{min:2.92,max:2.92,mean:2.92}},envoyProxy:{memory:{min:36.73,max:36.73,mean:36.73},cpu:{min:91.9,max:91.9,mean:91.9}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5542.3,totalRequests:166274,latency:{min:365,mean:6463,max:81305,pstdev:12357,percentiles:{p50:2890,p75:4545,p80:5116,p90:7585,p95:49043,p99:55365,p999:64399}},resources:{envoyGateway:{memory:{min:146.11,max:146.11,mean:146.11},cpu:{min:15.2,max:15.2,mean:15.2}},envoyProxy:{memory:{min:58.91,max:58.91,mean:58.91},cpu:{min:124.96,max:124.96,mean:124.96}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5407.59,totalRequests:162228,latency:{min:363,mean:6396,max:88236,pstdev:12419,percentiles:{p50:2816,p75:4450,p80:5012,p90:7485,p95:48762,p99:55666,p999:67457}},resources:{envoyGateway:{memory:{min:154.01,max:154.01,mean:154.01},cpu:{min:27.98,max:27.98,mean:27.98}},envoyProxy:{memory:{min:81.11,max:81.11,mean:81.11},cpu:{min:158.85,max:158.85,mean:158.85}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5247.32,totalRequests:157425,latency:{min:374,mean:6604,max:106491,pstdev:12903,percentiles:{p50:2880,p75:4514,p80:5089,p90:7639,p95:49453,p99:58750,p999:71970}},resources:{envoyGateway:{memory:{min:183.52,max:183.52,mean:183.52},cpu:{min:61.85,max:61.85,mean:61.85}},envoyProxy:{memory:{min:129.33,max:129.33,mean:129.33},cpu:{min:198.98,max:198.98,mean:198.98}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5668.05,totalRequests:170048,latency:{min:366,mean:6143,max:73551,pstdev:12054,percentiles:{p50:2726,p75:4279,p80:4799,p90:7123,p95:48988,p99:54704,p999:58812}},resources:{envoyGateway:{memory:{min:138.57,max:138.57,mean:138.57},cpu:{min:115.2,max:115.2,mean:115.2}},envoyProxy:{memory:{min:121.88,max:121.88,mean:121.88},cpu:{min:361.89,max:361.89,mean:361.89}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5579.16,totalRequests:167375,latency:{min:353,mean:6190,max:87244,pstdev:12028,percentiles:{p50:2803,p75:4357,p80:4889,p90:7110,p95:48482,p99:54329,p999:59113}},resources:{envoyGateway:{memory:{min:132.83,max:132.83,mean:132.83},cpu:{min:114.63,max:114.63,mean:114.63}},envoyProxy:{memory:{min:122.16,max:122.16,mean:122.16},cpu:{min:331.47,max:331.47,mean:331.47}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5581.7,totalRequests:167453,latency:{min:377,mean:6525,max:72724,pstdev:12286,percentiles:{p50:2929,p75:4641,p80:5219,p90:7915,p95:48732,p99:54810,p999:61450}},resources:{envoyGateway:{memory:{min:133.91,max:133.91,mean:133.91},cpu:{min:113.44,max:113.44,mean:113.44}},envoyProxy:{memory:{min:122.02,max:122.02,mean:122.02},cpu:{min:300.79,max:300.79,mean:300.79}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5447.53,totalRequests:163426,latency:{min:368,mean:6540,max:88612,pstdev:12451,percentiles:{p50:2904,p75:4546,p80:5128,p90:7692,p95:49305,p99:55422,p999:64116}},resources:{envoyGateway:{memory:{min:145.28,max:145.28,mean:145.28},cpu:{min:102.98,max:102.98,mean:102.98}},envoyProxy:{memory:{min:129.93,max:129.93,mean:129.93},cpu:{min:268.5,max:268.5,mean:268.5}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5437.61,totalRequests:163132,latency:{min:383,mean:6527,max:102998,pstdev:12600,percentiles:{p50:2827,p75:4470,p80:5070,p90:7714,p95:49029,p99:56436,p999:66684}},resources:{envoyGateway:{memory:{min:164.07,max:164.07,mean:164.07},cpu:{min:91.64,max:91.64,mean:91.64}},envoyProxy:{memory:{min:129.93,max:129.93,mean:129.93},cpu:{min:235.61,max:235.61,mean:235.61}}},poolOverflow:362,upstreamConnections:38}]},woe={metadata:{version:"1.2.4",runId:"1.2.4-1750189826764",date:"2024-12-13",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.4",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.4/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5982.38,totalRequests:179472,latency:{min:349,mean:6011,max:83349,pstdev:11061,percentiles:{p50:2829,p75:4477,p80:5098,p90:7842,p95:43255,p99:53350,p999:59346}},resources:{envoyGateway:{memory:{min:110.54,max:110.54,mean:110.54},cpu:{min:.77,max:.77,mean:.77}},envoyProxy:{memory:{min:24.25,max:24.25,mean:24.25},cpu:{min:30.44,max:30.44,mean:30.44}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5917.46,totalRequests:177524,latency:{min:361,mean:6164,max:70909,pstdev:11809,percentiles:{p50:2843,p75:4366,p80:4876,p90:7109,p95:47788,p99:54372,p999:59932}},resources:{envoyGateway:{memory:{min:119.89,max:119.89,mean:119.89},cpu:{min:1.61,max:1.61,mean:1.61}},envoyProxy:{memory:{min:30.43,max:30.43,mean:30.43},cpu:{min:61.13,max:61.13,mean:61.13}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5930.25,totalRequests:177910,latency:{min:398,mean:6027,max:92041,pstdev:11760,percentiles:{p50:2652,p75:4198,p80:4747,p90:7135,p95:47513,p99:55083,p999:61542}},resources:{envoyGateway:{memory:{min:129.08,max:129.08,mean:129.08},cpu:{min:3.05,max:3.05,mean:3.05}},envoyProxy:{memory:{min:36.57,max:36.57,mean:36.57},cpu:{min:92.01,max:92.01,mean:92.01}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5612.73,totalRequests:168382,latency:{min:368,mean:6361,max:108085,pstdev:12378,percentiles:{p50:2817,p75:4383,p80:4929,p90:7356,p95:48795,p99:56586,p999:68399}},resources:{envoyGateway:{memory:{min:140.63,max:140.63,mean:140.63},cpu:{min:15.35,max:15.35,mean:15.35}},envoyProxy:{memory:{min:58.77,max:58.77,mean:58.77},cpu:{min:125.44,max:125.44,mean:125.44}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5799.01,totalRequests:173974,latency:{min:378,mean:6166,max:97771,pstdev:11964,percentiles:{p50:2818,p75:4251,p80:4742,p90:6922,p95:47560,p99:55281,p999:68112}},resources:{envoyGateway:{memory:{min:152.5,max:152.5,mean:152.5},cpu:{min:28.38,max:28.38,mean:28.38}},envoyProxy:{memory:{min:78.94,max:78.94,mean:78.94},cpu:{min:159.67,max:159.67,mean:159.67}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5474.43,totalRequests:164233,latency:{min:386,mean:6534,max:109809,pstdev:12936,percentiles:{p50:2656,p75:4361,p80:5013,p90:8074,p95:49516,p99:60213,p999:73678}},resources:{envoyGateway:{memory:{min:183.98,max:183.98,mean:183.98},cpu:{min:62.29,max:62.29,mean:62.29}},envoyProxy:{memory:{min:131.16,max:131.16,mean:131.16},cpu:{min:200.46,max:200.46,mean:200.46}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5816.39,totalRequests:174492,latency:{min:340,mean:6131,max:85209,pstdev:11868,percentiles:{p50:2730,p75:4264,p80:4816,p90:7369,p95:47933,p99:54984,p999:61820}},resources:{envoyGateway:{memory:{min:195,max:195,mean:195},cpu:{min:116.53,max:116.53,mean:116.53}},envoyProxy:{memory:{min:120.37,max:120.37,mean:120.37},cpu:{min:363.61,max:363.61,mean:363.61}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5848.8,totalRequests:175470,latency:{min:372,mean:6285,max:77979,pstdev:12136,percentiles:{p50:2760,p75:4334,p80:4906,p90:7492,p95:48553,p99:55732,p999:63037}},resources:{envoyGateway:{memory:{min:205.2,max:205.2,mean:205.2},cpu:{min:115.88,max:115.88,mean:115.88}},envoyProxy:{memory:{min:120.26,max:120.26,mean:120.26},cpu:{min:333.06,max:333.06,mean:333.06}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5897.55,totalRequests:176927,latency:{min:365,mean:6054,max:78041,pstdev:11770,percentiles:{p50:2761,p75:4253,p80:4770,p90:6934,p95:47628,p99:54935,p999:62232}},resources:{envoyGateway:{memory:{min:201.78,max:201.78,mean:201.78},cpu:{min:114.84,max:114.84,mean:114.84}},envoyProxy:{memory:{min:122.23,max:122.23,mean:122.23},cpu:{min:302.66,max:302.66,mean:302.66}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5764.69,totalRequests:172941,latency:{min:376,mean:6540,max:87822,pstdev:12528,percentiles:{p50:2847,p75:4511,p80:5118,p90:7907,p95:49336,p99:56737,p999:66895}},resources:{envoyGateway:{memory:{min:237.36,max:237.36,mean:237.36},cpu:{min:104.3,max:104.3,mean:104.3}},envoyProxy:{memory:{min:131.2,max:131.2,mean:131.2},cpu:{min:270.32,max:270.32,mean:270.32}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5709.58,totalRequests:171290,latency:{min:350,mean:6172,max:93663,pstdev:11978,percentiles:{p50:2864,p75:4285,p80:4782,p90:6980,p95:47689,p99:55115,p999:67440}},resources:{envoyGateway:{memory:{min:158.41,max:158.41,mean:158.41},cpu:{min:92.69,max:92.69,mean:92.69}},envoyProxy:{memory:{min:131.32,max:131.32,mean:131.32},cpu:{min:236.85,max:236.85,mean:236.85}}},poolOverflow:363,upstreamConnections:37}]},boe={metadata:{version:"1.2.3",runId:"1.2.3-1750189826763",date:"2024-12-02",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.3",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.3/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:6222.52,totalRequests:186680,latency:{min:349,mean:6177,max:70643,pstdev:11464,percentiles:{p50:2879,p75:4520,p80:5082,p90:7684,p95:46213,p99:54147,p999:60250}},resources:{envoyGateway:{memory:{min:110.2,max:110.2,mean:110.2},cpu:{min:.68,max:.68,mean:.68}},envoyProxy:{memory:{min:25.89,max:25.89,mean:25.89},cpu:{min:30.33,max:30.33,mean:30.33}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:6170.38,totalRequests:185112,latency:{min:370,mean:5770,max:72798,pstdev:11358,percentiles:{p50:2657,p75:4054,p80:4535,p90:6537,p95:46632,p99:53751,p999:59174}},resources:{envoyGateway:{memory:{min:122.48,max:122.48,mean:122.48},cpu:{min:1.42,max:1.42,mean:1.42}},envoyProxy:{memory:{min:32.05,max:32.05,mean:32.05},cpu:{min:60.88,max:60.88,mean:60.88}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:6085.51,totalRequests:182567,latency:{min:365,mean:5863,max:84721,pstdev:11540,percentiles:{p50:2605,p75:4127,p80:4662,p90:7048,p95:46759,p99:54556,p999:61736}},resources:{envoyGateway:{memory:{min:131.06,max:131.06,mean:131.06},cpu:{min:2.8,max:2.8,mean:2.8}},envoyProxy:{memory:{min:38.19,max:38.19,mean:38.19},cpu:{min:91.64,max:91.64,mean:91.64}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:6032.83,totalRequests:180985,latency:{min:362,mean:6079,max:79421,pstdev:11806,percentiles:{p50:2707,p75:4272,p80:4832,p90:7327,p95:47441,p99:54743,p999:65808}},resources:{envoyGateway:{memory:{min:142.98,max:142.98,mean:142.98},cpu:{min:15.04,max:15.04,mean:15.04}},envoyProxy:{memory:{min:60.39,max:60.39,mean:60.39},cpu:{min:124.47,max:124.47,mean:124.47}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5773.98,totalRequests:173223,latency:{min:344,mean:6360,max:79355,pstdev:12473,percentiles:{p50:2707,p75:4299,p80:4888,p90:7643,p95:49309,p99:57382,p999:68141}},resources:{envoyGateway:{memory:{min:160.84,max:160.84,mean:160.84},cpu:{min:27.66,max:27.66,mean:27.66}},envoyProxy:{memory:{min:80.58,max:80.58,mean:80.58},cpu:{min:156.97,max:156.97,mean:156.97}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5776.08,totalRequests:173283,latency:{min:361,mean:6194,max:89714,pstdev:12241,percentiles:{p50:2703,p75:4241,p80:4791,p90:7296,p95:47755,p99:58537,p999:70676}},resources:{envoyGateway:{memory:{min:186.49,max:186.49,mean:186.49},cpu:{min:61.39,max:61.39,mean:61.39}},envoyProxy:{memory:{min:130.82,max:130.82,mean:130.82},cpu:{min:196.17,max:196.17,mean:196.17}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:6128.86,totalRequests:183866,latency:{min:351,mean:5824,max:73633,pstdev:11470,percentiles:{p50:2578,p75:4038,p80:4570,p90:6823,p95:46927,p99:54392,p999:60213}},resources:{envoyGateway:{memory:{min:218.36,max:218.36,mean:218.36},cpu:{min:114.4,max:114.4,mean:114.4}},envoyProxy:{memory:{min:118.8,max:118.8,mean:118.8},cpu:{min:358.03,max:358.03,mean:358.03}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:6065.62,totalRequests:181969,latency:{min:378,mean:5877,max:82526,pstdev:11643,percentiles:{p50:2588,p75:4048,p80:4587,p90:6876,p95:47429,p99:54710,p999:62111}},resources:{envoyGateway:{memory:{min:222.42,max:222.42,mean:222.42},cpu:{min:113.86,max:113.86,mean:113.86}},envoyProxy:{memory:{min:119.23,max:119.23,mean:119.23},cpu:{min:327.6,max:327.6,mean:327.6}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:6045.71,totalRequests:181360,latency:{min:388,mean:6024,max:80216,pstdev:11596,percentiles:{p50:2802,p75:4253,p80:4761,p90:6951,p95:47183,p99:54065,p999:59455}},resources:{envoyGateway:{memory:{min:142.37,max:142.37,mean:142.37},cpu:{min:112.77,max:112.77,mean:112.77}},envoyProxy:{memory:{min:119.54,max:119.54,mean:119.54},cpu:{min:296.95,max:296.95,mean:296.95}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5956.66,totalRequests:178700,latency:{min:379,mean:6153,max:85528,pstdev:12032,percentiles:{p50:2640,p75:4362,p80:4961,p90:7625,p95:47886,p99:55965,p999:65695}},resources:{envoyGateway:{memory:{min:148.72,max:148.72,mean:148.72},cpu:{min:102.49,max:102.49,mean:102.49}},envoyProxy:{memory:{min:130.98,max:130.98,mean:130.98},cpu:{min:264.96,max:264.96,mean:264.96}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5958.61,totalRequests:178765,latency:{min:362,mean:5967,max:83480,pstdev:11671,percentiles:{p50:2714,p75:4156,p80:4658,p90:6839,p95:46901,p99:54601,p999:65464}},resources:{envoyGateway:{memory:{min:164.87,max:164.87,mean:164.87},cpu:{min:91.25,max:91.25,mean:91.25}},envoyProxy:{memory:{min:130.98,max:130.98,mean:130.98},cpu:{min:231.98,max:231.98,mean:231.98}}},poolOverflow:363,upstreamConnections:37}]},Poe={metadata:{version:"1.2.2",runId:"1.2.2-1750189826761",date:"2024-11-28",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.2",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.2/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:6249.74,totalRequests:187495,latency:{min:333,mean:5937,max:75075,pstdev:10895,percentiles:{p50:2840,p75:4501,p80:5077,p90:7640,p95:41435,p99:53243,p999:58066}},resources:{envoyGateway:{memory:{min:106.26,max:106.26,mean:106.26},cpu:{min:.73,max:.73,mean:.73}},envoyProxy:{memory:{min:26.09,max:26.09,mean:26.09},cpu:{min:30.41,max:30.41,mean:30.41}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:6140.58,totalRequests:184222,latency:{min:373,mean:5812,max:79740,pstdev:11513,percentiles:{p50:2590,p75:4071,p80:4583,p90:6812,p95:46811,p99:54179,p999:61151}},resources:{envoyGateway:{memory:{min:131.24,max:131.24,mean:131.24},cpu:{min:1.34,max:1.34,mean:1.34}},envoyProxy:{memory:{min:32.26,max:32.26,mean:32.26},cpu:{min:60.74,max:60.74,mean:60.74}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:6149.72,totalRequests:184494,latency:{min:366,mean:5813,max:89935,pstdev:11434,percentiles:{p50:2637,p75:4073,p80:4561,p90:6680,p95:46505,p99:53872,p999:60348}},resources:{envoyGateway:{memory:{min:133.05,max:133.05,mean:133.05},cpu:{min:2.75,max:2.75,mean:2.75}},envoyProxy:{memory:{min:38.38,max:38.38,mean:38.38},cpu:{min:91.57,max:91.57,mean:91.57}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:6076.78,totalRequests:182305,latency:{min:361,mean:6031,max:87502,pstdev:11819,percentiles:{p50:2675,p75:4203,p80:4751,p90:7118,p95:47147,p99:55078,p999:67106}},resources:{envoyGateway:{memory:{min:153.62,max:153.62,mean:153.62},cpu:{min:14.77,max:14.77,mean:14.77}},envoyProxy:{memory:{min:60.45,max:60.45,mean:60.45},cpu:{min:124.04,max:124.04,mean:124.04}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5928.89,totalRequests:177873,latency:{min:336,mean:6196,max:85643,pstdev:12032,percentiles:{p50:2803,p75:4338,p80:4852,p90:7143,p95:47726,p99:55619,p999:67780}},resources:{envoyGateway:{memory:{min:145.79,max:145.79,mean:145.79},cpu:{min:27.17,max:27.17,mean:27.17}},envoyProxy:{memory:{min:80.64,max:80.64,mean:80.64},cpu:{min:157.65,max:157.65,mean:157.65}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5843.64,totalRequests:175314,latency:{min:339,mean:6452,max:101912,pstdev:12482,percentiles:{p50:2790,p75:4509,p80:5136,p90:8012,p95:47702,p99:58542,p999:74014}},resources:{envoyGateway:{memory:{min:173.7,max:173.7,mean:173.7},cpu:{min:61.12,max:61.12,mean:61.12}},envoyProxy:{memory:{min:130.89,max:130.89,mean:130.89},cpu:{min:195.74,max:195.74,mean:195.74}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:6096.9,totalRequests:182907,latency:{min:365,mean:5840,max:99377,pstdev:11326,percentiles:{p50:2735,p75:4188,p80:4674,p90:6654,p95:46344,p99:53444,p999:58591}},resources:{envoyGateway:{memory:{min:230.99,max:230.99,mean:230.99},cpu:{min:114.22,max:114.22,mean:114.22}},envoyProxy:{memory:{min:121.56,max:121.56,mean:121.56},cpu:{min:358.71,max:358.71,mean:358.71}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:6151.76,totalRequests:184556,latency:{min:376,mean:5805,max:84885,pstdev:11572,percentiles:{p50:2520,p75:3969,p80:4498,p90:6842,p95:46837,p99:54761,p999:62570}},resources:{envoyGateway:{memory:{min:233.63,max:233.63,mean:233.63},cpu:{min:113.65,max:113.65,mean:113.65}},envoyProxy:{memory:{min:121.86,max:121.86,mean:121.86},cpu:{min:328.3,max:328.3,mean:328.3}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:6157.53,totalRequests:184729,latency:{min:379,mean:5964,max:75988,pstdev:11502,percentiles:{p50:2743,p75:4204,p80:4695,p90:6941,p95:46497,p99:54179,p999:62005}},resources:{envoyGateway:{memory:{min:141.36,max:141.36,mean:141.36},cpu:{min:112.67,max:112.67,mean:112.67}},envoyProxy:{memory:{min:122.04,max:122.04,mean:122.04},cpu:{min:297.59,max:297.59,mean:297.59}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:6060.5,totalRequests:181818,latency:{min:360,mean:6022,max:92483,pstdev:11657,percentiles:{p50:2685,p75:4265,p80:4811,p90:7313,p95:46495,p99:54495,p999:64219}},resources:{envoyGateway:{memory:{min:150.71,max:150.71,mean:150.71},cpu:{min:102.4,max:102.4,mean:102.4}},envoyProxy:{memory:{min:130.9,max:130.9,mean:130.9},cpu:{min:265.29,max:265.29,mean:265.29}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5966.83,totalRequests:179008,latency:{min:343,mean:5962,max:102715,pstdev:11763,percentiles:{p50:2668,p75:4125,p80:4633,p90:6863,p95:46741,p99:54929,p999:67018}},resources:{envoyGateway:{memory:{min:162.62,max:162.62,mean:162.62},cpu:{min:91.03,max:91.03,mean:91.03}},envoyProxy:{memory:{min:131.05,max:131.05,mean:131.05},cpu:{min:232.55,max:232.55,mean:232.55}}},poolOverflow:363,upstreamConnections:37}]},Soe={metadata:{version:"1.2.1",runId:"1.2.1-1750189826759",date:"2024-11-07",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.1",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.1/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:6110.53,totalRequests:183316,latency:{min:337,mean:6415,max:92647,pstdev:11350,percentiles:{p50:3107,p75:5037,p80:5711,p90:8782,p95:45103,p99:53942,p999:60790}},resources:{envoyGateway:{memory:{min:111.18,max:111.18,mean:111.18},cpu:{min:.79,max:.79,mean:.79}},envoyProxy:{memory:{min:24.25,max:24.25,mean:24.25},cpu:{min:30.62,max:30.62,mean:30.62}}},poolOverflow:359,upstreamConnections:41},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:6011.43,totalRequests:180343,latency:{min:377,mean:6244,max:72724,pstdev:11870,percentiles:{p50:2835,p75:4408,p80:4939,p90:7412,p95:47611,p99:54769,p999:61716}},resources:{envoyGateway:{memory:{min:116.45,max:116.45,mean:116.45},cpu:{min:1.59,max:1.59,mean:1.59}},envoyProxy:{memory:{min:30.41,max:30.41,mean:30.41},cpu:{min:61.42,max:61.42,mean:61.42}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5961.65,totalRequests:178854,latency:{min:370,mean:5979,max:81096,pstdev:11586,percentiles:{p50:2705,p75:4189,p80:4699,p90:7030,p95:46987,p99:54095,p999:61222}},resources:{envoyGateway:{memory:{min:128.02,max:128.02,mean:128.02},cpu:{min:3.04,max:3.04,mean:3.04}},envoyProxy:{memory:{min:36.55,max:36.55,mean:36.55},cpu:{min:92.49,max:92.49,mean:92.49}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5921.3,totalRequests:177639,latency:{min:376,mean:6208,max:100098,pstdev:11901,percentiles:{p50:2770,p75:4413,p80:5009,p90:7647,p95:46934,p99:55273,p999:68509}},resources:{envoyGateway:{memory:{min:144.93,max:144.93,mean:144.93},cpu:{min:15.11,max:15.11,mean:15.11}},envoyProxy:{memory:{min:58.75,max:58.75,mean:58.75},cpu:{min:125.9,max:125.9,mean:125.9}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5876.24,totalRequests:176290,latency:{min:365,mean:6248,max:80101,pstdev:12134,percentiles:{p50:2759,p75:4348,p80:4922,p90:7422,p95:47855,p99:56041,p999:67145}},resources:{envoyGateway:{memory:{min:155.94,max:155.94,mean:155.94},cpu:{min:27.72,max:27.72,mean:27.72}},envoyProxy:{memory:{min:80.95,max:80.95,mean:80.95},cpu:{min:159.58,max:159.58,mean:159.58}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5744.1,totalRequests:172323,latency:{min:380,mean:6374,max:100245,pstdev:12424,percentiles:{p50:2775,p75:4337,p80:4899,p90:7542,p95:47699,p99:59248,p999:71761}},resources:{envoyGateway:{memory:{min:184.41,max:184.41,mean:184.41},cpu:{min:61.5,max:61.5,mean:61.5}},envoyProxy:{memory:{min:129.19,max:129.19,mean:129.19},cpu:{min:199.75,max:199.75,mean:199.75}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:6028.67,totalRequests:180864,latency:{min:361,mean:5919,max:75882,pstdev:11427,percentiles:{p50:2729,p75:4244,p80:4783,p90:6934,p95:46608,p99:53897,p999:59893}},resources:{envoyGateway:{memory:{min:229.41,max:229.41,mean:229.41},cpu:{min:114.16,max:114.16,mean:114.16}},envoyProxy:{memory:{min:121.69,max:121.69,mean:121.69},cpu:{min:363.78,max:363.78,mean:363.78}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:6093.21,totalRequests:182799,latency:{min:364,mean:5859,max:73809,pstdev:11454,percentiles:{p50:2634,p75:4171,p80:4702,p90:7002,p95:46548,p99:53979,p999:62793}},resources:{envoyGateway:{memory:{min:236.91,max:236.91,mean:236.91},cpu:{min:113.6,max:113.6,mean:113.6}},envoyProxy:{memory:{min:122.3,max:122.3,mean:122.3},cpu:{min:333.22,max:333.22,mean:333.22}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:6057.56,totalRequests:181727,latency:{min:366,mean:5898,max:83623,pstdev:11560,percentiles:{p50:2628,p75:4159,p80:4695,p90:6996,p95:46897,p99:54495,p999:61849}},resources:{envoyGateway:{memory:{min:143.71,max:143.71,mean:143.71},cpu:{min:112.58,max:112.58,mean:112.58}},envoyProxy:{memory:{min:121.84,max:121.84,mean:121.84},cpu:{min:302.58,max:302.58,mean:302.58}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5899.93,totalRequests:176998,latency:{min:329,mean:6218,max:88965,pstdev:12007,percentiles:{p50:2818,p75:4388,p80:4920,p90:7204,p95:47871,p99:55517,p999:65886}},resources:{envoyGateway:{memory:{min:154.24,max:154.24,mean:154.24},cpu:{min:102.3,max:102.3,mean:102.3}},envoyProxy:{memory:{min:129.34,max:129.34,mean:129.34},cpu:{min:270.21,max:270.21,mean:270.21}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5862.2,totalRequests:175868,latency:{min:359,mean:6106,max:107622,pstdev:11962,percentiles:{p50:2648,p75:4343,p80:4933,p90:7617,p95:47341,p99:55508,p999:69103}},resources:{envoyGateway:{memory:{min:153.72,max:153.72,mean:153.72},cpu:{min:91.07,max:91.07,mean:91.07}},envoyProxy:{memory:{min:129.18,max:129.18,mean:129.18},cpu:{min:237.04,max:237.04,mean:237.04}}},poolOverflow:363,upstreamConnections:37}]},Ooe={metadata:{version:"1.2.0",runId:"1.2.0-1750189826754",date:"2024-11-06",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.2.0",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.2.0/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:6006.27,totalRequests:180191,latency:{min:359,mean:6202,max:74358,pstdev:11262,percentiles:{p50:2976,p75:4719,p80:5330,p90:8159,p95:44003,p99:54190,p999:59357}},resources:{envoyGateway:{memory:{min:113.66,max:113.66,mean:113.66},cpu:{min:.75,max:.75,mean:.75}},envoyProxy:{memory:{min:25.55,max:25.55,mean:25.55},cpu:{min:30.41,max:30.41,mean:30.41}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5988.15,totalRequests:179650,latency:{min:366,mean:6093,max:82386,pstdev:11670,percentiles:{p50:2838,p75:4283,p80:4798,p90:6972,p95:47523,p99:54038,p999:59813}},resources:{envoyGateway:{memory:{min:119.16,max:119.16,mean:119.16},cpu:{min:1.53,max:1.53,mean:1.53}},envoyProxy:{memory:{min:31.71,max:31.71,mean:31.71},cpu:{min:61.03,max:61.03,mean:61.03}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5930.13,totalRequests:177905,latency:{min:374,mean:6182,max:93884,pstdev:11844,percentiles:{p50:2812,p75:4420,p80:4967,p90:7339,p95:47618,p99:54661,p999:61865}},resources:{envoyGateway:{memory:{min:123.35,max:123.35,mean:123.35},cpu:{min:2.89,max:2.89,mean:2.89}},envoyProxy:{memory:{min:37.86,max:37.86,mean:37.86},cpu:{min:91.87,max:91.87,mean:91.87}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5713.27,totalRequests:171401,latency:{min:384,mean:6405,max:105545,pstdev:12313,percentiles:{p50:2832,p75:4450,p80:5057,p90:7773,p95:48732,p99:56084,p999:66220}},resources:{envoyGateway:{memory:{min:150.92,max:150.92,mean:150.92},cpu:{min:15.08,max:15.08,mean:15.08}},envoyProxy:{memory:{min:60.05,max:60.05,mean:60.05},cpu:{min:125.09,max:125.09,mean:125.09}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5689.03,totalRequests:170675,latency:{min:365,mean:5939,max:98639,pstdev:11585,percentiles:{p50:2715,p75:4258,p80:4772,p90:7055,p95:45996,p99:54368,p999:68464}},resources:{envoyGateway:{memory:{min:156.97,max:156.97,mean:156.97},cpu:{min:27.62,max:27.62,mean:27.62}},envoyProxy:{memory:{min:82.26,max:82.26,mean:82.26},cpu:{min:158.86,max:158.86,mean:158.86}}},poolOverflow:365,upstreamConnections:35},{testName:"scale-up-httproutes-1000",routes:1e3,routesPerHostname:1,phase:"scaling-up",throughput:5407.08,totalRequests:162220,latency:{min:371,mean:6424,max:131579,pstdev:12473,percentiles:{p50:2692,p75:4503,p80:5177,p90:8264,p95:47540,p99:58488,p999:72720}},resources:{envoyGateway:{memory:{min:215.93,max:215.93,mean:215.93},cpu:{min:61.56,max:61.56,mean:61.56}},envoyProxy:{memory:{min:130.5,max:130.5,mean:130.5},cpu:{min:197.45,max:197.45,mean:197.45}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5905.23,totalRequests:177157,latency:{min:396,mean:6205,max:92979,pstdev:12065,percentiles:{p50:2713,p75:4262,p80:4832,p90:7415,p95:48306,p99:55736,p999:63318}},resources:{envoyGateway:{memory:{min:233.71,max:233.71,mean:233.71},cpu:{min:114.53,max:114.53,mean:114.53}},envoyProxy:{memory:{min:118.94,max:118.94,mean:118.94},cpu:{min:360.39,max:360.39,mean:360.39}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5793.19,totalRequests:173800,latency:{min:385,mean:6501,max:86843,pstdev:12452,percentiles:{p50:2744,p75:4568,p80:5227,p90:8494,p95:48965,p99:56467,p999:65382}},resources:{envoyGateway:{memory:{min:202.08,max:202.08,mean:202.08},cpu:{min:113.96,max:113.96,mean:113.96}},envoyProxy:{memory:{min:121.24,max:121.24,mean:121.24},cpu:{min:329.91,max:329.91,mean:329.91}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5783.8,totalRequests:173514,latency:{min:385,mean:6179,max:85405,pstdev:11889,percentiles:{p50:2707,p75:4462,p80:5071,p90:7760,p95:47423,p99:54994,p999:64567}},resources:{envoyGateway:{memory:{min:144.59,max:144.59,mean:144.59},cpu:{min:112.93,max:112.93,mean:112.93}},envoyProxy:{memory:{min:121.22,max:121.22,mean:121.22},cpu:{min:299.28,max:299.28,mean:299.28}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5804.51,totalRequests:174139,latency:{min:384,mean:6147,max:82661,pstdev:11892,percentiles:{p50:2807,p75:4284,p80:4791,p90:7039,p95:47558,p99:55162,p999:66723}},resources:{envoyGateway:{memory:{min:160.23,max:160.23,mean:160.23},cpu:{min:102.85,max:102.85,mean:102.85}},envoyProxy:{memory:{min:130.62,max:130.62,mean:130.62},cpu:{min:267.09,max:267.09,mean:267.09}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-down",throughput:5804.87,totalRequests:174149,latency:{min:368,mean:5828,max:106119,pstdev:11561,percentiles:{p50:2647,p75:4078,p80:4553,p90:6656,p95:46047,p99:55025,p999:66633}},resources:{envoyGateway:{memory:{min:155.88,max:155.88,mean:155.88},cpu:{min:91.5,max:91.5,mean:91.5}},envoyProxy:{memory:{min:130.61,max:130.61,mean:130.61},cpu:{min:234.47,max:234.47,mean:234.47}}},poolOverflow:365,upstreamConnections:35}]},koe={metadata:{version:"1.1.4",runId:"1.1.4-1750190329987",date:"2024-12-13",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.1.4",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.1.4/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5971.73,totalRequests:179152,latency:{min:354,mean:6247,max:97591,pstdev:11371,percentiles:{p50:2948,p75:4694,p80:5352,p90:8420,p95:44519,p99:54398,p999:62244}},resources:{envoyGateway:{memory:{min:82.79,max:82.79,mean:82.79},cpu:{min:1.5000000000000002,max:1.5000000000000002,mean:1.5000000000000002}},envoyProxy:{memory:{min:24.11,max:24.11,mean:24.11},cpu:{min:101.53333333333335,max:101.53333333333335,mean:101.53333333333335}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5957.22,totalRequests:178717,latency:{min:385,mean:6145,max:93360,pstdev:11838,percentiles:{p50:2751,p75:4352,p80:4922,p90:7436,p95:47679,p99:54878,p999:62414}},resources:{envoyGateway:{memory:{min:103.29,max:103.29,mean:103.29},cpu:{min:6.800000000000001,max:6.800000000000001,mean:6.800000000000001}},envoyProxy:{memory:{min:32.29,max:32.29,mean:32.29},cpu:{min:204,max:204,mean:204}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5896.53,totalRequests:176899,latency:{min:360,mean:6376,max:93995,pstdev:12077,percentiles:{p50:2905,p75:4542,p80:5112,p90:7699,p95:48019,p99:55472,p999:63707}},resources:{envoyGateway:{memory:{min:151.7,max:151.7,mean:151.7},cpu:{min:32.599999999999994,max:32.599999999999994,mean:32.599999999999994}},envoyProxy:{memory:{min:46.46,max:46.46,mean:46.46},cpu:{min:309.8,max:309.8,mean:309.8}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5885.1,totalRequests:176553,latency:{min:370,mean:6378,max:102457,pstdev:12096,percentiles:{p50:2834,p75:4508,p80:5093,p90:7817,p95:48099,p99:55572,p999:64329}},resources:{envoyGateway:{memory:{min:763.38,max:763.38,mean:763.38},cpu:{min:605.6333333333333,max:605.6333333333333,mean:605.6333333333333}},envoyProxy:{memory:{min:150.81,max:150.81,mean:150.81},cpu:{min:494.66666666666674,max:494.66666666666674,mean:494.66666666666674}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5642.37,totalRequests:169274,latency:{min:314,mean:6430,max:135987,pstdev:12216,percentiles:{p50:2801,p75:4489,p80:5175,p90:8673,p95:46690,p99:56358,p999:73814}},resources:{envoyGateway:{memory:{min:1603.28,max:1603.28,mean:1603.28},cpu:{min:41.63333333333333,max:41.63333333333333,mean:41.63333333333333}},envoyProxy:{memory:{min:289.75,max:289.75,mean:289.75},cpu:{min:727.1666666666666,max:727.1666666666666,mean:727.1666666666666}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5993.8,totalRequests:179822,latency:{min:381,mean:5965,max:84086,pstdev:11660,percentiles:{p50:2666,p75:4169,p80:4715,p90:6995,p95:47218,p99:54775,p999:61200}},resources:{envoyGateway:{memory:{min:117.16,max:117.16,mean:117.16},cpu:{min:518.0333333333333,max:518.0333333333333,mean:518.0333333333333}},envoyProxy:{memory:{min:278.72,max:278.72,mean:278.72},cpu:{min:1205.8,max:1205.8,mean:1205.8}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5947.1,totalRequests:178412,latency:{min:346,mean:5809,max:76681,pstdev:11155,percentiles:{p50:2707,p75:4164,p80:4684,p90:6938,p95:45451,p99:53227,p999:59303}},resources:{envoyGateway:{memory:{min:123.23,max:123.23,mean:123.23},cpu:{min:512.5333333333333,max:512.5333333333333,mean:512.5333333333333}},envoyProxy:{memory:{min:291.25,max:291.25,mean:291.25},cpu:{min:1103.6333333333334,max:1103.6333333333334,mean:1103.6333333333334}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:4884.46,totalRequests:146534,latency:{min:362,mean:7148,max:137863,pstdev:11237,percentiles:{p50:3375,p75:6266,p80:7621,p90:17699,p95:31992,p99:59576,p999:83935}},resources:{envoyGateway:{memory:{min:338.12,max:338.12,mean:338.12},cpu:{min:488.1333333333333,max:488.1333333333333,mean:488.1333333333333}},envoyProxy:{memory:{min:294.41,max:294.41,mean:294.41},cpu:{min:993,max:993,mean:993}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5737.62,totalRequests:172120,latency:{min:345,mean:6192,max:117288,pstdev:11530,percentiles:{p50:2822,p75:4535,p80:5155,p90:8463,p95:42631,p99:55584,p999:70291}},resources:{envoyGateway:{memory:{min:679.73,max:679.73,mean:679.73},cpu:{min:18.76666666666667,max:18.76666666666667,mean:18.76666666666667}},envoyProxy:{memory:{min:285.93,max:285.93,mean:285.93},cpu:{min:833.4333333333334,max:833.4333333333334,mean:833.4333333333334}}},poolOverflow:363,upstreamConnections:37}]},Coe={metadata:{version:"1.1.3",runId:"1.1.3-1750190329984",date:"2024-11-04",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.1.3",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.1.3/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:6070.53,totalRequests:182116,latency:{min:353,mean:6263,max:87117,pstdev:11274,percentiles:{p50:2994,p75:4783,p80:5441,p90:8506,p95:44029,p99:54130,p999:59291}},resources:{envoyGateway:{memory:{min:85.16,max:85.16,mean:85.16},cpu:{min:1.3666666666666665,max:1.3666666666666665,mean:1.3666666666666665}},envoyProxy:{memory:{min:26,max:26,mean:26},cpu:{min:101.43333333333334,max:101.43333333333334,mean:101.43333333333334}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5999.45,totalRequests:179985,latency:{min:368,mean:6117,max:76840,pstdev:11675,percentiles:{p50:2803,p75:4339,p80:4903,p90:7204,p95:47128,p99:54317,p999:60557}},resources:{envoyGateway:{memory:{min:102.18,max:102.18,mean:102.18},cpu:{min:7.033333333333333,max:7.033333333333333,mean:7.033333333333333}},envoyProxy:{memory:{min:34.18,max:34.18,mean:34.18},cpu:{min:203.56666666666666,max:203.56666666666666,mean:203.56666666666666}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:6053.06,totalRequests:181592,latency:{min:390,mean:5889,max:94826,pstdev:11642,percentiles:{p50:2599,p75:4075,p80:4583,p90:6945,p95:47128,p99:54872,p999:62212}},resources:{envoyGateway:{memory:{min:151.34,max:151.34,mean:151.34},cpu:{min:31.633333333333336,max:31.633333333333336,mean:31.633333333333336}},envoyProxy:{memory:{min:48.34,max:48.34,mean:48.34},cpu:{min:309.06666666666666,max:309.06666666666666,mean:309.06666666666666}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5895.77,totalRequests:176873,latency:{min:378,mean:6220,max:100757,pstdev:12033,percentiles:{p50:2723,p75:4341,p80:4931,p90:7644,p95:47828,p99:55521,p999:67383}},resources:{envoyGateway:{memory:{min:687.16,max:687.16,mean:687.16},cpu:{min:578.2333333333332,max:578.2333333333332,mean:578.2333333333332}},envoyProxy:{memory:{min:150.69,max:150.69,mean:150.69},cpu:{min:488.20000000000005,max:488.20000000000005,mean:488.20000000000005}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5818.09,totalRequests:174543,latency:{min:356,mean:6109,max:107913,pstdev:11285,percentiles:{p50:2832,p75:4538,p80:5163,p90:8134,p95:42639,p99:54546,p999:67543}},resources:{envoyGateway:{memory:{min:1329.35,max:1329.35,mean:1329.35},cpu:{min:61.33333333333333,max:61.33333333333333,mean:61.33333333333333}},envoyProxy:{memory:{min:303.55,max:303.55,mean:303.55},cpu:{min:731.4333333333334,max:731.4333333333334,mean:731.4333333333334}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5996.79,totalRequests:179907,latency:{min:367,mean:6127,max:77688,pstdev:11777,percentiles:{p50:2745,p75:4292,p80:4835,p90:7317,p95:47597,p99:54706,p999:61323}},resources:{envoyGateway:{memory:{min:125.62,max:125.62,mean:125.62},cpu:{min:525.5,max:525.5,mean:525.5}},envoyProxy:{memory:{min:289.68,max:289.68,mean:289.68},cpu:{min:1239.6,max:1239.6,mean:1239.6}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5732.14,totalRequests:171956,latency:{min:336,mean:4876,max:76279,pstdev:10072,percentiles:{p50:2290,p75:3444,p80:3846,p90:5476,p95:24546,p99:51763,p999:57040}},resources:{envoyGateway:{memory:{min:132.64,max:132.64,mean:132.64},cpu:{min:520.8333333333333,max:520.8333333333333,mean:520.8333333333333}},envoyProxy:{memory:{min:291.68,max:291.68,mean:291.68},cpu:{min:1138.0333333333335,max:1138.0333333333335,mean:1138.0333333333335}}},poolOverflow:371,upstreamConnections:29},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:5311.34,totalRequests:159332,latency:{min:362,mean:6219,max:144302,pstdev:11133,percentiles:{p50:2808,p75:4741,p80:5579,p90:10689,p95:37480,p99:53626,p999:73998}},resources:{envoyGateway:{memory:{min:179.01,max:179.01,mean:179.01},cpu:{min:497.9666666666666,max:497.9666666666666,mean:497.9666666666666}},envoyProxy:{memory:{min:298.77,max:298.77,mean:298.77},cpu:{min:1034.033333333333,max:1034.033333333333,mean:1034.033333333333}}},poolOverflow:365,upstreamConnections:35},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5633.85,totalRequests:169021,latency:{min:369,mean:6266,max:106577,pstdev:11127,percentiles:{p50:2908,p75:4793,p80:5511,p90:9568,p95:36169,p99:54472,p999:68628}},resources:{envoyGateway:{memory:{min:487.12,max:487.12,mean:487.12},cpu:{min:27.133333333333336,max:27.133333333333336,mean:27.133333333333336}},envoyProxy:{memory:{min:299.84,max:299.84,mean:299.84},cpu:{min:837.8666666666668,max:837.8666666666668,mean:837.8666666666668}}},poolOverflow:363,upstreamConnections:37}]},_oe={metadata:{version:"1.1.2",runId:"1.1.2-1750190329982",date:"2024-09-24",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.1.2",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.1.2/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:5649.72,totalRequests:169493,latency:{min:356,mean:6560,max:72347,pstdev:11796,percentiles:{p50:3044,p75:5011,p80:5690,p90:8937,p95:46880,p99:54560,p999:59316}},resources:{envoyGateway:{memory:{min:86.72,max:86.72,mean:86.72},cpu:{min:1.5666666666666667,max:1.5666666666666667,mean:1.5666666666666667}},envoyProxy:{memory:{min:24.26,max:24.26,mean:24.26},cpu:{min:101.63333333333333,max:101.63333333333333,mean:101.63333333333333}}},poolOverflow:360,upstreamConnections:40},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:5594.14,totalRequests:167826,latency:{min:377,mean:6225,max:71024,pstdev:11980,percentiles:{p50:2812,p75:4378,p80:4905,p90:7214,p95:48160,p99:54513,p999:59850}},resources:{envoyGateway:{memory:{min:106.76,max:106.76,mean:106.76},cpu:{min:7.7,max:7.7,mean:7.7}},envoyProxy:{memory:{min:32.43,max:32.43,mean:32.43},cpu:{min:204.53333333333333,max:204.53333333333333,mean:204.53333333333333}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:5523.65,totalRequests:165711,latency:{min:379,mean:6455,max:77815,pstdev:12317,percentiles:{p50:2868,p75:4510,p80:5086,p90:7645,p95:49127,p99:55048,p999:61448}},resources:{envoyGateway:{memory:{min:149.03,max:149.03,mean:149.03},cpu:{min:29.333333333333332,max:29.333333333333332,mean:29.333333333333332}},envoyProxy:{memory:{min:46.61,max:46.61,mean:46.61},cpu:{min:309.7666666666667,max:309.7666666666667,mean:309.7666666666667}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5554.01,totalRequests:166624,latency:{min:357,mean:6246,max:99753,pstdev:12051,percentiles:{p50:2841,p75:4416,p80:4960,p90:7177,p95:48371,p99:54800,p999:63275}},resources:{envoyGateway:{memory:{min:375.51,max:375.51,mean:375.51},cpu:{min:615.1,max:615.1,mean:615.1}},envoyProxy:{memory:{min:152.97,max:152.97,mean:152.97},cpu:{min:492.4666666666667,max:492.4666666666667,mean:492.4666666666667}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5475.04,totalRequests:164256,latency:{min:365,mean:6546,max:95539,pstdev:12646,percentiles:{p50:2834,p75:4484,p80:5075,p90:7891,p95:49397,p99:56231,p999:67629}},resources:{envoyGateway:{memory:{min:1340.6,max:1340.6,mean:1340.6},cpu:{min:38.333333333333336,max:38.333333333333336,mean:38.333333333333336}},envoyProxy:{memory:{min:281.71,max:281.71,mean:281.71},cpu:{min:718.0666666666666,max:718.0666666666666,mean:718.0666666666666}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5657.68,totalRequests:169734,latency:{min:355,mean:7320,max:87252,pstdev:13162,percentiles:{p50:3307,p75:5218,p80:5885,p90:9252,p95:50671,p99:56332,p999:62406}},resources:{envoyGateway:{memory:{min:371.29,max:371.29,mean:371.29},cpu:{min:578.7666666666667,max:578.7666666666667,mean:578.7666666666667}},envoyProxy:{memory:{min:273.57,max:273.57,mean:273.57},cpu:{min:1224,max:1224,mean:1224}}},poolOverflow:356,upstreamConnections:44},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5587.8,totalRequests:167636,latency:{min:356,mean:6346,max:68095,pstdev:12097,percentiles:{p50:2890,p75:4509,p80:5041,p90:7495,p95:48560,p99:54345,p999:60436}},resources:{envoyGateway:{memory:{min:342.8,max:342.8,mean:342.8},cpu:{min:573.3000000000001,max:573.3000000000001,mean:573.3000000000001}},envoyProxy:{memory:{min:274.84,max:274.84,mean:274.84},cpu:{min:1122.2333333333333,max:1122.2333333333333,mean:1122.2333333333333}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:4098.74,totalRequests:122966,latency:{min:347,mean:3660,max:99745,pstdev:7371,percentiles:{p50:1658,p75:2735,p80:3147,p90:5306,p95:18030,p99:42778,p999:60440}},resources:{envoyGateway:{memory:{min:433.01,max:433.01,mean:433.01},cpu:{min:486.96666666666664,max:486.96666666666664,mean:486.96666666666664}},envoyProxy:{memory:{min:281.27,max:281.27,mean:281.27},cpu:{min:996.0333333333333,max:996.0333333333333,mean:996.0333333333333}}},poolOverflow:384,upstreamConnections:16},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5306.46,totalRequests:159194,latency:{min:360,mean:6884,max:106807,pstdev:12850,percentiles:{p50:2983,p75:4873,p80:5573,p90:9215,p95:49102,p99:58521,p999:70213}},resources:{envoyGateway:{memory:{min:232.76,max:232.76,mean:232.76},cpu:{min:16.299999999999997,max:16.299999999999997,mean:16.299999999999997}},envoyProxy:{memory:{min:282.68,max:282.68,mean:282.68},cpu:{min:825.4333333333333,max:825.4333333333333,mean:825.4333333333333}}},poolOverflow:361,upstreamConnections:39}]},Aoe={metadata:{version:"1.1.1",runId:"1.1.1-1750190329981",date:"2024-09-12",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.1.1",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.1.1/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:6258.99,totalRequests:187774,latency:{min:333,mean:5933,max:71266,pstdev:10816,percentiles:{p50:2898,p75:4513,p80:5099,p90:7661,p95:40607,p99:53149,p999:58789}},resources:{envoyGateway:{memory:{min:84.04,max:84.04,mean:84.04},cpu:{min:1.5333333333333334,max:1.5333333333333334,mean:1.5333333333333334}},envoyProxy:{memory:{min:25.57,max:25.57,mean:25.57},cpu:{min:101.33333333333331,max:101.33333333333331,mean:101.33333333333331}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:6118.28,totalRequests:183552,latency:{min:359,mean:5828,max:81178,pstdev:11437,percentiles:{p50:2646,p75:4049,p80:4561,p90:6666,p95:46723,p99:53913,p999:60094}},resources:{envoyGateway:{memory:{min:101.33,max:101.33,mean:101.33},cpu:{min:6.8999999999999995,max:6.8999999999999995,mean:6.8999999999999995}},envoyProxy:{memory:{min:33.75,max:33.75,mean:33.75},cpu:{min:203.6,max:203.6,mean:203.6}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:6045.36,totalRequests:181361,latency:{min:377,mean:5908,max:88063,pstdev:11539,percentiles:{p50:2659,p75:4143,p80:4675,p90:6863,p95:46733,p99:54091,p999:63211}},resources:{envoyGateway:{memory:{min:115.34,max:115.34,mean:115.34},cpu:{min:32.800000000000004,max:32.800000000000004,mean:32.800000000000004}},envoyProxy:{memory:{min:47.79,max:47.79,mean:47.79},cpu:{min:308.96666666666664,max:308.96666666666664,mean:308.96666666666664}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5881.18,totalRequests:176436,latency:{min:327,mean:6376,max:130330,pstdev:12048,percentiles:{p50:2843,p75:4548,p80:5181,p90:8168,p95:47517,p99:55470,p999:66041}},resources:{envoyGateway:{memory:{min:608.65,max:608.65,mean:608.65},cpu:{min:625.6666666666666,max:625.6666666666666,mean:625.6666666666666}},envoyProxy:{memory:{min:152.29,max:152.29,mean:152.29},cpu:{min:492.5666666666667,max:492.5666666666667,mean:492.5666666666667}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5875.93,totalRequests:176280,latency:{min:381,mean:6086,max:86654,pstdev:11932,percentiles:{p50:2704,p75:4181,p80:4727,p90:7177,p95:47091,p99:55814,p999:68415}},resources:{envoyGateway:{memory:{min:1308.52,max:1308.52,mean:1308.52},cpu:{min:39.49999999999999,max:39.49999999999999,mean:39.49999999999999}},envoyProxy:{memory:{min:283.24,max:283.24,mean:283.24},cpu:{min:707.1333333333332,max:707.1333333333332,mean:707.1333333333332}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:6077.56,totalRequests:182327,latency:{min:362,mean:6197,max:92372,pstdev:11720,percentiles:{p50:2824,p75:4438,p80:5e3,p90:7503,p95:47110,p99:54169,p999:60426}},resources:{envoyGateway:{memory:{min:121.13,max:121.13,mean:121.13},cpu:{min:544.5333333333334,max:544.5333333333334,mean:544.5333333333334}},envoyProxy:{memory:{min:277.29,max:277.29,mean:277.29},cpu:{min:1206.4666666666665,max:1206.4666666666665,mean:1206.4666666666665}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:6047.67,totalRequests:181438,latency:{min:370,mean:6062,max:92168,pstdev:11743,percentiles:{p50:2733,p75:4268,p80:4801,p90:7226,p95:47095,p99:54886,p999:61917}},resources:{envoyGateway:{memory:{min:126.22,max:126.22,mean:126.22},cpu:{min:538.8333333333334,max:538.8333333333334,mean:538.8333333333334}},envoyProxy:{memory:{min:285.62,max:285.62,mean:285.62},cpu:{min:1104.3000000000002,max:1104.3000000000002,mean:1104.3000000000002}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:4654.8,totalRequests:139655,latency:{min:336,mean:4675,max:107409,pstdev:9843,percentiles:{p50:2048,p75:3459,p80:4015,p90:6531,p95:19138,p99:57518,p999:73269}},resources:{envoyGateway:{memory:{min:480.36,max:480.36,mean:480.36},cpu:{min:377,max:377,mean:377}},envoyProxy:{memory:{min:289.51,max:289.51,mean:289.51},cpu:{min:969.2333333333332,max:969.2333333333332,mean:969.2333333333332}}},poolOverflow:377,upstreamConnections:23},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5676.68,totalRequests:170301,latency:{min:376,mean:6244,max:153264,pstdev:11618,percentiles:{p50:2797,p75:4618,p80:5324,p90:8807,p95:40232,p99:56496,p999:73506}},resources:{envoyGateway:{memory:{min:436.58,max:436.58,mean:436.58},cpu:{min:28.066666666666666,max:28.066666666666666,mean:28.066666666666666}},envoyProxy:{memory:{min:283.47,max:283.47,mean:283.47},cpu:{min:813.3666666666667,max:813.3666666666667,mean:813.3666666666667}}},poolOverflow:363,upstreamConnections:37}]},Eoe={metadata:{version:"1.1.0",runId:"1.1.0-1750190329977",date:"2024-07-23",environment:"GitHub CI",description:"Benchmark report for Envoy Gateway 1.1.0",downloadUrl:"https://github.com/envoyproxy/gateway/releases/download/v1.1.0/benchmark_report.zip",testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"}},results:[{testName:"scale-up-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-up",throughput:6181.97,totalRequests:185463,latency:{min:362,mean:5902,max:73084,pstdev:11039,percentiles:{p50:2765,p75:4364,p80:4935,p90:7504,p95:41244,p99:53929,p999:61147}},resources:{envoyGateway:{memory:{min:86.2,max:86.2,mean:86.2},cpu:{min:1.4333333333333333,max:1.4333333333333333,mean:1.4333333333333333}},envoyProxy:{memory:{min:24.09,max:24.09,mean:24.09},cpu:{min:101.36666666666667,max:101.36666666666667,mean:101.36666666666667}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-up",throughput:6103.97,totalRequests:183121,latency:{min:354,mean:5852,max:75943,pstdev:11382,percentiles:{p50:2681,p75:4083,p80:4601,p90:6825,p95:46262,p99:53862,p999:60651}},resources:{envoyGateway:{memory:{min:100.14,max:100.14,mean:100.14},cpu:{min:7.233333333333333,max:7.233333333333333,mean:7.233333333333333}},envoyProxy:{memory:{min:32.25,max:32.25,mean:32.25},cpu:{min:203.6,max:203.6,mean:203.6}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-up",throughput:6074.87,totalRequests:182247,latency:{min:371,mean:5877,max:94171,pstdev:11399,percentiles:{p50:2732,p75:4191,p80:4677,p90:6801,p95:46481,p99:53866,p999:61036}},resources:{envoyGateway:{memory:{min:114.69,max:114.69,mean:114.69},cpu:{min:30,max:30,mean:30}},envoyProxy:{memory:{min:46.42,max:46.42,mean:46.42},cpu:{min:308.70000000000005,max:308.70000000000005,mean:308.70000000000005}}},poolOverflow:363,upstreamConnections:37},{testName:"scale-up-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-up",throughput:5993.57,totalRequests:179811,latency:{min:368,mean:6123,max:96571,pstdev:11831,percentiles:{p50:2753,p75:4261,p80:4799,p90:7247,p95:47368,p99:55248,p999:66586}},resources:{envoyGateway:{memory:{min:762.8,max:762.8,mean:762.8},cpu:{min:600.8666666666667,max:600.8666666666667,mean:600.8666666666667}},envoyProxy:{memory:{min:152.78,max:152.78,mean:152.78},cpu:{min:488.70000000000005,max:488.70000000000005,mean:488.70000000000005}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-up-httproutes-500",routes:500,routesPerHostname:1,phase:"scaling-up",throughput:5812.8,totalRequests:174396,latency:{min:388,mean:6310,max:95719,pstdev:12246,percentiles:{p50:2729,p75:4334,p80:4905,p90:7607,p95:47970,p99:56811,p999:68165}},resources:{envoyGateway:{memory:{min:1593.56,max:1593.56,mean:1593.56},cpu:{min:36.56666666666667,max:36.56666666666667,mean:36.56666666666667}},envoyProxy:{memory:{min:289.79,max:289.79,mean:289.79},cpu:{min:715.6333333333333,max:715.6333333333333,mean:715.6333333333333}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-10",routes:10,routesPerHostname:1,phase:"scaling-down",throughput:5916.93,totalRequests:177517,latency:{min:385,mean:6352,max:82747,pstdev:12011,percentiles:{p50:2822,p75:4461,p80:5061,p90:7769,p95:47847,p99:55068,p999:62578}},resources:{envoyGateway:{memory:{min:124.07,max:124.07,mean:124.07},cpu:{min:519.6333333333332,max:519.6333333333332,mean:519.6333333333332}},envoyProxy:{memory:{min:284.03,max:284.03,mean:284.03},cpu:{min:1194.8666666666666,max:1194.8666666666666,mean:1194.8666666666666}}},poolOverflow:361,upstreamConnections:39},{testName:"scale-down-httproutes-50",routes:50,routesPerHostname:1,phase:"scaling-down",throughput:5876.35,totalRequests:176294,latency:{min:371,mean:6226,max:83869,pstdev:11816,percentiles:{p50:2770,p75:4396,p80:5e3,p90:7848,p95:47163,p99:55156,p999:64434}},resources:{envoyGateway:{memory:{min:131.62,max:131.62,mean:131.62},cpu:{min:513.6,max:513.6,mean:513.6}},envoyProxy:{memory:{min:291.73,max:291.73,mean:291.73},cpu:{min:1092.9333333333334,max:1092.9333333333334,mean:1092.9333333333334}}},poolOverflow:362,upstreamConnections:38},{testName:"scale-down-httproutes-100",routes:100,routesPerHostname:1,phase:"scaling-down",throughput:4820.43,totalRequests:144617,latency:{min:342,mean:7088,max:172031,pstdev:12642,percentiles:{p50:3156,p75:5480,p80:6443,p90:12007,p95:39675,p99:63842,p999:86532}},resources:{envoyGateway:{memory:{min:374.36,max:374.36,mean:374.36},cpu:{min:462.8,max:462.8,mean:462.8}},envoyProxy:{memory:{min:289.98,max:289.98,mean:289.98},cpu:{min:980.4,max:980.4,mean:980.4}}},poolOverflow:364,upstreamConnections:36},{testName:"scale-down-httproutes-300",routes:300,routesPerHostname:1,phase:"scaling-down",throughput:5746.71,totalRequests:172402,latency:{min:386,mean:6313,max:101535,pstdev:11578,percentiles:{p50:2875,p75:4579,p80:5222,p90:8617,p95:44341,p99:54996,p999:67178}},resources:{envoyGateway:{memory:{min:670.44,max:670.44,mean:670.44},cpu:{min:24.333333333333332,max:24.333333333333332,mean:24.333333333333332}},envoyProxy:{memory:{min:285.86,max:285.86,mean:285.86},cpu:{min:824.5666666666667,max:824.5666666666667,mean:824.5666666666667}}},poolOverflow:362,upstreamConnections:38}]},Gg=[foe,poe,coe,doe,moe,hoe,voe,yoe,goe,xoe,woe,boe,Poe,Soe,Ooe,koe,Coe,_oe,Aoe,Eoe],Toe=()=>Gg.map(e=>e.metadata.version),mi=Gg.find(e=>e.metadata.version==="1.4.1"),bt=(mi==null?void 0:mi.results)||[];mi!=null&&mi.metadata.testConfiguration;bt.length,bt.filter(e=>e.phase==="scaling-up").length,bt.filter(e=>e.phase==="scaling-down").length,bt.length>0&&Math.max(...bt.map(e=>e.routes)),bt.length>0&&Math.min(...bt.map(e=>e.routes)),bt.length>0&&bt.reduce((e,t)=>e+t.throughput,0)/bt.length,bt.length>0&&bt.reduce((e,t)=>e+t.latency.mean,0)/bt.length;bt.map(e=>({routes:e.routes,phase:e.phase,p50:e.latency.percentiles.p50/1e3,p75:e.latency.percentiles.p75/1e3,p90:e.latency.percentiles.p90/1e3,p95:e.latency.percentiles.p95/1e3,p99:e.latency.percentiles.p99/1e3,p999:e.latency.percentiles.p999/1e3}));bt.map(e=>({routes:e.routes,phase:e.phase,envoyGatewayMemory:e.resources.envoyGateway.memory.mean,envoyGatewayCpu:e.resources.envoyGateway.cpu.mean,envoyProxyMemory:e.resources.envoyProxy.memory.mean,envoyProxyCpu:e.resources.envoyProxy.cpu.mean}));bt.map(e=>({testName:e.testName,routes:e.routes,phase:e.phase,throughput:e.throughput,meanLatency:e.latency.mean/1e3,p95Latency:e.latency.percentiles.p95/1e3,totalMemory:e.resources.envoyGateway.memory.mean+e.resources.envoyProxy.memory.mean,totalCpu:e.resources.envoyGateway.cpu.mean+e.resources.envoyProxy.cpu.mean}));const joe=()=>{const e=Toe(),[t,n]=k.useState(e[0]||""),r=k.useMemo(()=>Gg.find(o=>o.metadata.version===t),[t]),a=k.useMemo(()=>{if(!r)return{benchmarkResults:[],testConfiguration:{rps:1e4,connections:100,duration:30,cpuLimit:"1000m",memoryLimit:"2000Mi"},performanceSummary:{totalTests:0,scaleUpTests:0,scaleDownTests:0,maxRoutes:0,minRoutes:0,avgThroughput:0,avgLatency:0},latencyPercentileComparison:[],resourceTrends:[],performanceMatrix:[],metadata:null};const o=r.results;return{benchmarkResults:o,testConfiguration:r.metadata.testConfiguration,performanceSummary:{totalTests:o.length,scaleUpTests:o.filter(i=>i.phase==="scaling-up").length,scaleDownTests:o.filter(i=>i.phase==="scaling-down").length,maxRoutes:o.length>0?Math.max(...o.map(i=>i.routes)):0,minRoutes:o.length>0?Math.min(...o.map(i=>i.routes)):0,avgThroughput:o.length>0?o.reduce((i,s)=>i+s.throughput,0)/o.length:0,avgLatency:o.length>0?o.reduce((i,s)=>i+s.latency.mean,0)/o.length:0},latencyPercentileComparison:o.map(i=>({routes:i.routes,phase:i.phase,p50:i.latency.percentiles.p50/1e3,p75:i.latency.percentiles.p75/1e3,p90:i.latency.percentiles.p90/1e3,p95:i.latency.percentiles.p95/1e3,p99:i.latency.percentiles.p99/1e3,p999:i.latency.percentiles.p999/1e3})),resourceTrends:o.map(i=>({routes:i.routes,phase:i.phase,envoyGatewayMemory:i.resources.envoyGateway.memory.mean,envoyGatewayCpu:i.resources.envoyGateway.cpu.mean,envoyProxyMemory:i.resources.envoyProxy.memory.mean,envoyProxyCpu:i.resources.envoyProxy.cpu.mean})),performanceMatrix:o.map(i=>({testName:i.testName,routes:i.routes,phase:i.phase,throughput:i.throughput,meanLatency:i.latency.mean/1e3,p95Latency:i.latency.percentiles.p95/1e3,totalMemory:i.resources.envoyGateway.memory.mean+i.resources.envoyProxy.memory.mean,totalCpu:i.resources.envoyGateway.cpu.mean+i.resources.envoyProxy.cpu.mean})),metadata:r.metadata}},[r]);return{selectedVersion:t,setSelectedVersion:n,availableVersions:e,...a}},Noe=({apiBase:e="https://envoy-gateway-benchmark-report.netlify.app/api",initialVersion:t,theme:n="light",containerClassName:r="",features:a={header:!1,versionSelector:!0,summaryCards:!0,tabs:["overview","latency","resources"]}})=>{var i,s,l,u,p,c;const o=joe();return k.useEffect(()=>{n==="dark"?document.documentElement.classList.add("dark"):document.documentElement.classList.remove("dark")},[n]),k.useEffect(()=>{const f=document.createElement("style");return f.textContent=`
      .benchmark-dashboard [data-radix-popper-content-wrapper] {
        z-index: 9999 !important;
      }
      .benchmark-dashboard .relative.z-50 {
        z-index: 9999 !important;
      }
    `,document.head.appendChild(f),()=>{document.head.removeChild(f)}},[]),b.jsxs("div",{className:`benchmark-dashboard ${n} ${r}`,"data-theme":n,children:[a.header&&b.jsxs("div",{className:"mb-8",children:[b.jsx("h2",{className:"text-3xl font-bold",children:"Performance Benchmark Report Explorer"}),b.jsx("p",{className:"text-xl text-gray-600 dark:text-gray-300",children:"Detailed performance analysis"})]}),a.versionSelector&&b.jsx("div",{className:"mb-6",children:b.jsx("div",{className:"bg-white dark:bg-gray-800 rounded-xl shadow-lg p-4 w-full",children:b.jsx(BT,{selectedVersion:o.selectedVersion,availableVersions:o.availableVersions,onVersionChange:o.setSelectedVersion,metadata:o.metadata})})}),a.summaryCards&&o.performanceSummary&&b.jsx("div",{className:"mb-8",children:b.jsx(HT,{performanceSummary:o.performanceSummary,benchmarkResults:o.benchmarkResults})}),a.tabs&&a.tabs.length>0&&b.jsx("div",{className:"bg-white dark:bg-gray-800 rounded-2xl shadow-lg overflow-hidden",children:b.jsxs(uoe,{defaultValue:a.tabs[0],className:"w-full",children:[b.jsxs(WS,{className:`grid w-full ${a.tabs.length===1?"grid-cols-1":a.tabs.length===2?"grid-cols-2":"grid-cols-3"} bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 h-auto p-0 rounded-none`,children:[((i=a.tabs)==null?void 0:i.includes("overview"))&&b.jsx(Mc,{value:"overview",className:"data-[state=active]:bg-gradient-to-r data-[state=active]:from-purple-600 data-[state=active]:to-indigo-600 data-[state=active]:text-white data-[state=active]:shadow-lg data-[state=active]:border-b-2 data-[state=active]:border-purple-600 hover:bg-gray-50 dark:hover:bg-gray-700 text-sm sm:text-base py-4 px-6 rounded-t-lg border-b-2 border-transparent transition-all duration-200 font-medium",children:"Overview"}),((s=a.tabs)==null?void 0:s.includes("latency"))&&b.jsx(Mc,{value:"latency",className:"data-[state=active]:bg-gradient-to-r data-[state=active]:from-purple-600 data-[state=active]:to-indigo-600 data-[state=active]:text-white data-[state=active]:shadow-lg data-[state=active]:border-b-2 data-[state=active]:border-purple-600 hover:bg-gray-50 dark:hover:bg-gray-700 text-sm sm:text-base py-4 px-6 rounded-t-lg border-b-2 border-transparent transition-all duration-200 font-medium",children:"Request RTT Analysis"}),((l=a.tabs)==null?void 0:l.includes("resources"))&&b.jsx(Mc,{value:"resources",className:"data-[state=active]:bg-gradient-to-r data-[state=active]:from-purple-600 data-[state=active]:to-indigo-600 data-[state=active]:text-white data-[state=active]:shadow-lg data-[state=active]:border-b-2 data-[state=active]:border-purple-600 hover:bg-gray-50 dark:hover:bg-gray-700 text-sm sm:text-base py-4 px-6 rounded-t-lg border-b-2 border-transparent transition-all duration-200 font-medium",children:"Resource Usage"})]}),b.jsxs("div",{className:"p-6",children:[((u=a.tabs)==null?void 0:u.includes("overview"))&&b.jsx(Rc,{value:"overview",children:b.jsx(Hae,{performanceMatrix:o.performanceMatrix,benchmarkResults:o.benchmarkResults,testConfiguration:o.testConfiguration,performanceSummary:o.performanceSummary,latencyPercentileComparison:o.latencyPercentileComparison})}),((p=a.tabs)==null?void 0:p.includes("latency"))&&b.jsx(Rc,{value:"latency",children:b.jsx(Gae,{latencyPercentileComparison:o.latencyPercentileComparison,benchmarkResults:o.benchmarkResults})}),((c=a.tabs)==null?void 0:c.includes("resources"))&&b.jsx(Rc,{value:"resources",children:b.jsx(Uae,{resourceTrends:o.resourceTrends,benchmarkResults:o.benchmarkResults})})]})]})})]})};class qS{static async getAllCSS(){const t="shadow-dom-css";if(this.cssCache.has(t))return this.cssCache.get(t);try{let n="";const r=this.getBundledCSS();if(r){n=this.transformCSSForShadowDOM(r);const a=`
          ${this.getShadowDOMStyles()}
          ${n}
        `;return this.cssCache.set(t,a),a}else{const a=this.getFallbackCSS();return this.cssCache.set(t,a),a}}catch(n){return console.error("Error loading CSS for Shadow DOM:",n),this.getFallbackCSS()}}static getBundledCSS(){return window.__BENCHMARK_CSS__||null}static getCSSVariables(){return`
      /* CSS Custom Properties for theming */
      :host(.benchmark-dashboard) {
        --background: 0 0% 100%;
        --foreground: 222.2 84% 4.9%;
        --card: 0 0% 100%;
        --card-foreground: 222.2 84% 4.9%;
        --popover: 0 0% 100%;
        --popover-foreground: 222.2 84% 4.9%;
        --primary: 222.2 47.4% 11.2%;
        --primary-foreground: 210 40% 98%;
        --secondary: 210 40% 96.1%;
        --secondary-foreground: 222.2 47.4% 11.2%;
        --muted: 210 40% 96.1%;
        --muted-foreground: 215.4 16.3% 46.9%;
        --accent: 210 40% 96.1%;
        --accent-foreground: 222.2 47.4% 11.2%;
        --destructive: 0 84.2% 60.2%;
        --destructive-foreground: 210 40% 98%;
        --border: 214.3 31.8% 91.4%;
        --input: 214.3 31.8% 91.4%;
        --ring: 222.2 84% 4.9%;
        --radius: 0.5rem;
        --sidebar-background: 0 0% 98%;
        --sidebar-foreground: 240 5.3% 26.1%;
        --sidebar-primary: 240 5.9% 10%;
        --sidebar-primary-foreground: 0 0% 98%;
        --sidebar-accent: 240 4.8% 95.9%;
        --sidebar-accent-foreground: 240 5.9% 10%;
        --sidebar-border: 220 13% 91%;
        --sidebar-ring: 217.2 91.2% 59.8%;
      }

      :host(.benchmark-dashboard.dark) {
        --background: 222.2 84% 4.9%;
        --foreground: 210 40% 98%;
        --card: 222.2 84% 4.9%;
        --card-foreground: 210 40% 98%;
        --popover: 222.2 84% 4.9%;
        --popover-foreground: 210 40% 98%;
        --primary: 210 40% 98%;
        --primary-foreground: 222.2 47.4% 11.2%;
        --secondary: 217.2 32.6% 17.5%;
        --secondary-foreground: 210 40% 98%;
        --muted: 217.2 32.6% 17.5%;
        --muted-foreground: 215 20.2% 65.1%;
        --accent: 217.2 32.6% 17.5%;
        --accent-foreground: 210 40% 98%;
        --destructive: 0 62.8% 30.6%;
        --destructive-foreground: 210 40% 98%;
        --border: 217.2 32.6% 17.5%;
        --input: 217.2 32.6% 17.5%;
        --ring: 212.7 26.8% 83.9%;
        --sidebar-background: 240 5.9% 10%;
        --sidebar-foreground: 240 4.8% 95.9%;
        --sidebar-primary: 224.3 76.3% 48%;
        --sidebar-primary-foreground: 0 0% 100%;
        --sidebar-accent: 240 3.7% 15.9%;
        --sidebar-accent-foreground: 240 4.8% 95.9%;
        --sidebar-border: 240 3.7% 15.9%;
        --sidebar-ring: 217.2 91.2% 59.8%;
      }
    `}static getShadowDOMStyles(){return`
      /* Shadow DOM host styles */
      :host {
        all: initial;
        display: block;
        font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
        line-height: 1.5;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        color: hsl(var(--foreground));
        background: hsl(var(--background));
      }

      /* Reset all elements */
      * {
        box-sizing: border-box;
      }

      /* Shadow root container */
      #shadow-root {
        width: 100%;
        min-height: 100%;
        position: relative;
      }

      /* Ensure proper isolation */
      :host([hidden]) {
        display: none;
      }
    `}static transformCSSForShadowDOM(t){return t=t.replace(/:root\s*{([^}]*)}/g,":host {$1}"),t=t.replace(/\.dark\s*,\s*body\.dark\s*{([^}]*)}/g,":host(.dark) {$1}"),t=t.replace(/\.dark\s*{([^}]*)}/g,":host(.dark) {$1}"),t=t.replace(/body\.dark\s*{([^}]*)}/g,":host(.dark) {$1}"),t=t.replace(/body\s*{([^}]*)}/g,":host {$1}"),t}static getFallbackCSS(){return`
      ${this.getShadowDOMStyles()}
      ${this.getCSSVariables()}
      ${this.getTailwindFallback()}
    `}static getTailwindFallback(){return`
      /* Tailwind-like reset */
      *,
      ::before,
      ::after {
        box-sizing: border-box;
        border-width: 0;
        border-style: solid;
        border-color: hsl(var(--border));
      }

      /* Display utilities */
      .flex { display: flex !important; }
      .grid { display: grid !important; }
      .block { display: block !important; }
      .inline-block { display: inline-block !important; }
      .inline { display: inline !important; }
      .hidden { display: none !important; }
      .inline-flex { display: inline-flex !important; }

      /* Sizing utilities */
      .w-full { width: 100% !important; }
      .w-8 { width: 2rem !important; }
      .w-auto { width: auto !important; }
      .h-full { height: 100% !important; }
      .h-8 { height: 2rem !important; }
      .h-auto { height: auto !important; }
      .min-h-full { min-height: 100% !important; }

      /* Spacing utilities - Padding */
      .p-0 { padding: 0 !important; }
      .p-1 { padding: 0.25rem !important; }
      .p-2 { padding: 0.5rem !important; }
      .p-4 { padding: 1rem !important; }
      .p-6 { padding: 1.5rem !important; }
      .p-8 { padding: 2rem !important; }
      .px-3 { padding-left: 0.75rem !important; padding-right: 0.75rem !important; }
      .px-4 { padding-left: 1rem !important; padding-right: 1rem !important; }
      .px-6 { padding-left: 1.5rem !important; padding-right: 1.5rem !important; }
      .py-2 { padding-top: 0.5rem !important; padding-bottom: 0.5rem !important; }
      .py-4 { padding-top: 1rem !important; padding-bottom: 1rem !important; }
      .py-12 { padding-top: 3rem !important; padding-bottom: 3rem !important; }

      /* Spacing utilities - Margin */
      .m-0 { margin: 0 !important; }
      .m-1 { margin: 0.25rem !important; }
      .m-2 { margin: 0.5rem !important; }
      .m-4 { margin: 1rem !important; }
      .m-6 { margin: 1.5rem !important; }
      .m-8 { margin: 2rem !important; }
      .mb-2 { margin-bottom: 0.5rem !important; }
      .mb-4 { margin-bottom: 1rem !important; }
      .mb-5 { margin-bottom: 1.25rem !important; }
      .mb-6 { margin-bottom: 1.5rem !important; }
      .mb-8 { margin-bottom: 2rem !important; }
      .mt-4 { margin-top: 1rem !important; }
      .mt-5 { margin-top: 1.25rem !important; }
      .mr-3 { margin-right: 0.75rem !important; }

      /* Typography */
      .text-sm { font-size: 0.875rem !important; line-height: 1.25rem !important; }
      .text-base { font-size: 1rem !important; line-height: 1.5rem !important; }
      .text-lg { font-size: 1.125rem !important; line-height: 1.75rem !important; }
      .text-xl { font-size: 1.25rem !important; line-height: 1.75rem !important; }
      .text-2xl { font-size: 1.5rem !important; line-height: 2rem !important; }
      .text-3xl { font-size: 1.875rem !important; line-height: 2.25rem !important; }

      .font-medium { font-weight: 500 !important; }
      .font-semibold { font-weight: 600 !important; }
      .font-bold { font-weight: 700 !important; }

      /* Text colors */
      .text-white { color: #ffffff !important; }
      .text-gray-600 { color: hsl(var(--muted-foreground)) !important; }
      .text-gray-700 { color: hsl(var(--foreground)) !important; }
      .text-gray-800 { color: hsl(var(--foreground)) !important; }
      .text-gray-900 { color: hsl(var(--foreground)) !important; }

      /* Background colors */
      .bg-white { background-color: hsl(var(--background)) !important; }
      .bg-gray-50 { background-color: hsl(var(--muted)) !important; }
      .bg-gray-100 { background-color: hsl(var(--muted)) !important; }
      .bg-gray-700 { background-color: hsl(var(--muted)) !important; }
      .bg-gray-800 { background-color: hsl(var(--card)) !important; }
      .bg-gradient-to-r { background-image: linear-gradient(to right, var(--tw-gradient-stops)) !important; }

      /* Gradient colors */
      .from-purple-600 { --tw-gradient-from: #9333ea !important; --tw-gradient-to: rgb(147 51 234 / 0) !important; --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to) !important; }
      .to-indigo-600 { --tw-gradient-to: #4f46e5 !important; }

      /* Border utilities */
      .border { border-width: 1px !important; }
      .border-b { border-bottom-width: 1px !important; }
      .border-b-2 { border-bottom-width: 2px !important; }
      .border-t-lg { border-top-width: 4px !important; }
      .border-gray-200 { border-color: hsl(var(--border)) !important; }
      .border-gray-300 { border-color: hsl(var(--border)) !important; }
      .border-gray-700 { border-color: hsl(var(--border)) !important; }
      .border-blue-600 { border-color: #2563eb !important; }
      .border-purple-600 { border-color: #9333ea !important; }
      .border-transparent { border-color: transparent !important; }

      /* Border radius */
      .rounded { border-radius: 0.25rem !important; }
      .rounded-md { border-radius: 0.375rem !important; }
      .rounded-lg { border-radius: 0.5rem !important; }
      .rounded-xl { border-radius: 0.75rem !important; }
      .rounded-2xl { border-radius: 1rem !important; }
      .rounded-t-lg { border-top-left-radius: 0.5rem !important; border-top-right-radius: 0.5rem !important; }
      .rounded-full { border-radius: 9999px !important; }
      .rounded-none { border-radius: 0px !important; }

      /* Box shadow */
      .shadow { box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1) !important; }
      .shadow-lg { box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1) !important; }

      /* Grid utilities */
      .grid-cols-1 { grid-template-columns: repeat(1, minmax(0, 1fr)) !important; }
      .grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)) !important; }
      .grid-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)) !important; }

      .gap-1 { gap: 0.25rem !important; }
      .gap-2 { gap: 0.5rem !important; }
      .gap-4 { gap: 1rem !important; }

      /* Flexbox utilities */
      .items-center { align-items: center !important; }
      .items-start { align-items: flex-start !important; }
      .justify-center { justify-content: center !important; }
      .justify-between { justify-content: space-between !important; }
      .justify-start { justify-content: flex-start !important; }

      /* Position utilities */
      .relative { position: relative !important; }
      .absolute { position: absolute !important; }
      .z-50 { z-index: 50 !important; }

      /* Overflow utilities */
      .overflow-hidden { overflow: hidden !important; }
      .overflow-auto { overflow: auto !important; }

      /* Cursor utilities */
      .cursor-pointer { cursor: pointer !important; }

      /* Transition utilities */
      .transition-all { transition-property: all !important; transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1) !important; transition-duration: 150ms !important; }
      .duration-200 { transition-duration: 200ms !important; }

      /* Hover utilities */
      .hover\\:bg-gray-50:hover { background-color: hsl(var(--muted)) !important; }
      .hover\\:bg-gray-700:hover { background-color: hsl(var(--muted)) !important; }

      /* Animation utilities */
      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }
      .animate-spin {
        animation: spin 1s linear infinite !important;
      }

      /* Dark mode utilities */
      .dark\\:bg-gray-700 { background-color: hsl(var(--muted)) !important; }
      .dark\\:bg-gray-800 { background-color: hsl(var(--card)) !important; }
      .dark\\:border-gray-700 { border-color: hsl(var(--border)) !important; }
      .dark\\:text-gray-300 { color: hsl(var(--muted-foreground)) !important; }
      .dark\\:hover\\:bg-gray-700:hover { background-color: hsl(var(--muted)) !important; }

      /* Data attribute utilities for Radix components */
      [data-state="active"] {
        background: linear-gradient(to right, #9333ea, #4f46e5) !important;
        color: white !important;
        box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1) !important;
        border-bottom: 2px solid #9333ea !important;
      }

      /* White space utilities */
      .whitespace-nowrap { white-space: nowrap !important; }

      /* Text alignment */
      .text-center { text-align: center !important; }
      .text-left { text-align: left !important; }

      /* Responsive utilities (basic support) */
      @media (min-width: 640px) {
        .sm\\:text-base { font-size: 1rem !important; line-height: 1.5rem !important; }
      }
    `}}Qg(qS,"cssCache",new Map);var Ad=class{constructor(){this.listeners=new Set,this.subscribe=this.subscribe.bind(this)}subscribe(e){return this.listeners.add(e),this.onSubscribe(),()=>{this.listeners.delete(e),this.onUnsubscribe()}}hasListeners(){return this.listeners.size>0}onSubscribe(){}onUnsubscribe(){}},Ed=typeof window>"u"||"Deno"in globalThis;function An(){}function $oe(e,t){return typeof e=="function"?e(t):e}function Moe(e){return typeof e=="number"&&e>=0&&e!==1/0}function Roe(e,t){return Math.max(e+(t||0)-Date.now(),0)}function Pv(e,t){return typeof e=="function"?e(t):e}function Ioe(e,t){return typeof e=="function"?e(t):e}function k2(e,t){const{type:n="all",exact:r,fetchStatus:a,predicate:o,queryKey:i,stale:s}=e;if(i){if(r){if(t.queryHash!==Ug(i,t.options))return!1}else if(!mu(t.queryKey,i))return!1}if(n!=="all"){const l=t.isActive();if(n==="active"&&!l||n==="inactive"&&l)return!1}return!(typeof s=="boolean"&&t.isStale()!==s||a&&a!==t.state.fetchStatus||o&&!o(t))}function C2(e,t){const{exact:n,status:r,predicate:a,mutationKey:o}=e;if(o){if(!t.options.mutationKey)return!1;if(n){if(du(t.options.mutationKey)!==du(o))return!1}else if(!mu(t.options.mutationKey,o))return!1}return!(r&&t.state.status!==r||a&&!a(t))}function Ug(e,t){return((t==null?void 0:t.queryKeyHashFn)||du)(e)}function du(e){return JSON.stringify(e,(t,n)=>Sv(n)?Object.keys(n).sort().reduce((r,a)=>(r[a]=n[a],r),{}):n)}function mu(e,t){return e===t?!0:typeof e!=typeof t?!1:e&&t&&typeof e=="object"&&typeof t=="object"?Object.keys(t).every(n=>mu(e[n],t[n])):!1}function VS(e,t){if(e===t)return e;const n=_2(e)&&_2(t);if(n||Sv(e)&&Sv(t)){const r=n?e:Object.keys(e),a=r.length,o=n?t:Object.keys(t),i=o.length,s=n?[]:{},l=new Set(r);let u=0;for(let p=0;p<i;p++){const c=n?p:o[p];(!n&&l.has(c)||n)&&e[c]===void 0&&t[c]===void 0?(s[c]=void 0,u++):(s[c]=VS(e[c],t[c]),s[c]===e[c]&&e[c]!==void 0&&u++)}return a===i&&u===a?e:s}return t}function _2(e){return Array.isArray(e)&&e.length===Object.keys(e).length}function Sv(e){if(!A2(e))return!1;const t=e.constructor;if(t===void 0)return!0;const n=t.prototype;return!(!A2(n)||!n.hasOwnProperty("isPrototypeOf")||Object.getPrototypeOf(e)!==Object.prototype)}function A2(e){return Object.prototype.toString.call(e)==="[object Object]"}function Doe(e){return new Promise(t=>{setTimeout(t,e)})}function Loe(e,t,n){return typeof n.structuralSharing=="function"?n.structuralSharing(e,t):n.structuralSharing!==!1?VS(e,t):t}function Foe(e,t,n=0){const r=[...e,t];return n&&r.length>n?r.slice(1):r}function Boe(e,t,n=0){const r=[t,...e];return n&&r.length>n?r.slice(0,-1):r}var Wg=Symbol();function KS(e,t){return!e.queryFn&&(t!=null&&t.initialPromise)?()=>t.initialPromise:!e.queryFn||e.queryFn===Wg?()=>Promise.reject(new Error(`Missing queryFn: '${e.queryHash}'`)):e.queryFn}var eo,aa,hi,$2,zoe=($2=class extends Ad{constructor(){super();me(this,eo);me(this,aa);me(this,hi);ne(this,hi,t=>{if(!Ed&&window.addEventListener){const n=()=>t();return window.addEventListener("visibilitychange",n,!1),()=>{window.removeEventListener("visibilitychange",n)}}})}onSubscribe(){z(this,aa)||this.setEventListener(z(this,hi))}onUnsubscribe(){var t;this.hasListeners()||((t=z(this,aa))==null||t.call(this),ne(this,aa,void 0))}setEventListener(t){var n;ne(this,hi,t),(n=z(this,aa))==null||n.call(this),ne(this,aa,t(r=>{typeof r=="boolean"?this.setFocused(r):this.onFocus()}))}setFocused(t){z(this,eo)!==t&&(ne(this,eo,t),this.onFocus())}onFocus(){const t=this.isFocused();this.listeners.forEach(n=>{n(t)})}isFocused(){var t;return typeof z(this,eo)=="boolean"?z(this,eo):((t=globalThis.document)==null?void 0:t.visibilityState)!=="hidden"}},eo=new WeakMap,aa=new WeakMap,hi=new WeakMap,$2),XS=new zoe,vi,oa,yi,M2,Hoe=(M2=class extends Ad{constructor(){super();me(this,vi,!0);me(this,oa);me(this,yi);ne(this,yi,t=>{if(!Ed&&window.addEventListener){const n=()=>t(!0),r=()=>t(!1);return window.addEventListener("online",n,!1),window.addEventListener("offline",r,!1),()=>{window.removeEventListener("online",n),window.removeEventListener("offline",r)}}})}onSubscribe(){z(this,oa)||this.setEventListener(z(this,yi))}onUnsubscribe(){var t;this.hasListeners()||((t=z(this,oa))==null||t.call(this),ne(this,oa,void 0))}setEventListener(t){var n;ne(this,yi,t),(n=z(this,oa))==null||n.call(this),ne(this,oa,t(this.setOnline.bind(this)))}setOnline(t){z(this,vi)!==t&&(ne(this,vi,t),this.listeners.forEach(r=>{r(t)}))}isOnline(){return z(this,vi)}},vi=new WeakMap,oa=new WeakMap,yi=new WeakMap,M2),vf=new Hoe;function Goe(){let e,t;const n=new Promise((a,o)=>{e=a,t=o});n.status="pending",n.catch(()=>{});function r(a){Object.assign(n,a),delete n.resolve,delete n.reject}return n.resolve=a=>{r({status:"fulfilled",value:a}),e(a)},n.reject=a=>{r({status:"rejected",reason:a}),t(a)},n}function Uoe(e){return Math.min(1e3*2**e,3e4)}function YS(e){return(e??"online")==="online"?vf.isOnline():!0}var QS=class extends Error{constructor(e){super("CancelledError"),this.revert=e==null?void 0:e.revert,this.silent=e==null?void 0:e.silent}};function Wm(e){return e instanceof QS}function ZS(e){let t=!1,n=0,r=!1,a;const o=Goe(),i=d=>{var g;r||(f(new QS(d)),(g=e.abort)==null||g.call(e))},s=()=>{t=!0},l=()=>{t=!1},u=()=>XS.isFocused()&&(e.networkMode==="always"||vf.isOnline())&&e.canRun(),p=()=>YS(e.networkMode)&&e.canRun(),c=d=>{var g;r||(r=!0,(g=e.onSuccess)==null||g.call(e,d),a==null||a(),o.resolve(d))},f=d=>{var g;r||(r=!0,(g=e.onError)==null||g.call(e,d),a==null||a(),o.reject(d))},m=()=>new Promise(d=>{var g;a=v=>{(r||u())&&d(v)},(g=e.onPause)==null||g.call(e)}).then(()=>{var d;a=void 0,r||(d=e.onContinue)==null||d.call(e)}),h=()=>{if(r)return;let d;const g=n===0?e.initialPromise:void 0;try{d=g??e.fn()}catch(v){d=Promise.reject(v)}Promise.resolve(d).then(c).catch(v=>{var S;if(r)return;const y=e.retry??(Ed?0:3),x=e.retryDelay??Uoe,P=typeof x=="function"?x(n,v):x,w=y===!0||typeof y=="number"&&n<y||typeof y=="function"&&y(n,v);if(t||!w){f(v);return}n++,(S=e.onFail)==null||S.call(e,n,v),Doe(P).then(()=>u()?void 0:m()).then(()=>{t?f(v):h()})})};return{promise:o,cancel:i,continue:()=>(a==null||a(),o),cancelRetry:s,continueRetry:l,canStart:p,start:()=>(p()?h():m().then(h),o)}}var Woe=e=>setTimeout(e,0);function qoe(){let e=[],t=0,n=s=>{s()},r=s=>{s()},a=Woe;const o=s=>{t?e.push(s):a(()=>{n(s)})},i=()=>{const s=e;e=[],s.length&&a(()=>{r(()=>{s.forEach(l=>{n(l)})})})};return{batch:s=>{let l;t++;try{l=s()}finally{t--,t||i()}return l},batchCalls:s=>(...l)=>{o(()=>{s(...l)})},schedule:o,setNotifyFunction:s=>{n=s},setBatchNotifyFunction:s=>{r=s},setScheduler:s=>{a=s}}}var Nt=qoe(),to,R2,JS=(R2=class{constructor(){me(this,to)}destroy(){this.clearGcTimeout()}scheduleGc(){this.clearGcTimeout(),Moe(this.gcTime)&&ne(this,to,setTimeout(()=>{this.optionalRemove()},this.gcTime))}updateGcTime(e){this.gcTime=Math.max(this.gcTime||0,e??(Ed?1/0:5*60*1e3))}clearGcTimeout(){z(this,to)&&(clearTimeout(z(this,to)),ne(this,to,void 0))}},to=new WeakMap,R2),gi,no,cn,ro,Pt,hu,ao,En,xr,I2,Voe=(I2=class extends JS{constructor(t){super();me(this,En);me(this,gi);me(this,no);me(this,cn);me(this,ro);me(this,Pt);me(this,hu);me(this,ao);ne(this,ao,!1),ne(this,hu,t.defaultOptions),this.setOptions(t.options),this.observers=[],ne(this,ro,t.client),ne(this,cn,z(this,ro).getQueryCache()),this.queryKey=t.queryKey,this.queryHash=t.queryHash,ne(this,gi,Xoe(this.options)),this.state=t.state??z(this,gi),this.scheduleGc()}get meta(){return this.options.meta}get promise(){var t;return(t=z(this,Pt))==null?void 0:t.promise}setOptions(t){this.options={...z(this,hu),...t},this.updateGcTime(this.options.gcTime)}optionalRemove(){!this.observers.length&&this.state.fetchStatus==="idle"&&z(this,cn).remove(this)}setData(t,n){const r=Loe(this.state.data,t,this.options);return gt(this,En,xr).call(this,{data:r,type:"success",dataUpdatedAt:n==null?void 0:n.updatedAt,manual:n==null?void 0:n.manual}),r}setState(t,n){gt(this,En,xr).call(this,{type:"setState",state:t,setStateOptions:n})}cancel(t){var r,a;const n=(r=z(this,Pt))==null?void 0:r.promise;return(a=z(this,Pt))==null||a.cancel(t),n?n.then(An).catch(An):Promise.resolve()}destroy(){super.destroy(),this.cancel({silent:!0})}reset(){this.destroy(),this.setState(z(this,gi))}isActive(){return this.observers.some(t=>Ioe(t.options.enabled,this)!==!1)}isDisabled(){return this.getObserversCount()>0?!this.isActive():this.options.queryFn===Wg||this.state.dataUpdateCount+this.state.errorUpdateCount===0}isStatic(){return this.getObserversCount()>0?this.observers.some(t=>Pv(t.options.staleTime,this)==="static"):!1}isStale(){return this.getObserversCount()>0?this.observers.some(t=>t.getCurrentResult().isStale):this.state.data===void 0||this.state.isInvalidated}isStaleByTime(t=0){return this.state.data===void 0?!0:t==="static"?!1:this.state.isInvalidated?!0:!Roe(this.state.dataUpdatedAt,t)}onFocus(){var n;const t=this.observers.find(r=>r.shouldFetchOnWindowFocus());t==null||t.refetch({cancelRefetch:!1}),(n=z(this,Pt))==null||n.continue()}onOnline(){var n;const t=this.observers.find(r=>r.shouldFetchOnReconnect());t==null||t.refetch({cancelRefetch:!1}),(n=z(this,Pt))==null||n.continue()}addObserver(t){this.observers.includes(t)||(this.observers.push(t),this.clearGcTimeout(),z(this,cn).notify({type:"observerAdded",query:this,observer:t}))}removeObserver(t){this.observers.includes(t)&&(this.observers=this.observers.filter(n=>n!==t),this.observers.length||(z(this,Pt)&&(z(this,ao)?z(this,Pt).cancel({revert:!0}):z(this,Pt).cancelRetry()),this.scheduleGc()),z(this,cn).notify({type:"observerRemoved",query:this,observer:t}))}getObserversCount(){return this.observers.length}invalidate(){this.state.isInvalidated||gt(this,En,xr).call(this,{type:"invalidate"})}fetch(t,n){var u,p,c;if(this.state.fetchStatus!=="idle"){if(this.state.data!==void 0&&(n!=null&&n.cancelRefetch))this.cancel({silent:!0});else if(z(this,Pt))return z(this,Pt).continueRetry(),z(this,Pt).promise}if(t&&this.setOptions(t),!this.options.queryFn){const f=this.observers.find(m=>m.options.queryFn);f&&this.setOptions(f.options)}const r=new AbortController,a=f=>{Object.defineProperty(f,"signal",{enumerable:!0,get:()=>(ne(this,ao,!0),r.signal)})},o=()=>{const f=KS(this.options,n),h=(()=>{const d={client:z(this,ro),queryKey:this.queryKey,meta:this.meta};return a(d),d})();return ne(this,ao,!1),this.options.persister?this.options.persister(f,h,this):f(h)},s=(()=>{const f={fetchOptions:n,options:this.options,queryKey:this.queryKey,client:z(this,ro),state:this.state,fetchFn:o};return a(f),f})();(u=this.options.behavior)==null||u.onFetch(s,this),ne(this,no,this.state),(this.state.fetchStatus==="idle"||this.state.fetchMeta!==((p=s.fetchOptions)==null?void 0:p.meta))&&gt(this,En,xr).call(this,{type:"fetch",meta:(c=s.fetchOptions)==null?void 0:c.meta});const l=f=>{var m,h,d,g;Wm(f)&&f.silent||gt(this,En,xr).call(this,{type:"error",error:f}),Wm(f)||((h=(m=z(this,cn).config).onError)==null||h.call(m,f,this),(g=(d=z(this,cn).config).onSettled)==null||g.call(d,this.state.data,f,this)),this.scheduleGc()};return ne(this,Pt,ZS({initialPromise:n==null?void 0:n.initialPromise,fn:s.fetchFn,abort:r.abort.bind(r),onSuccess:f=>{var m,h,d,g;if(f===void 0){l(new Error(`${this.queryHash} data is undefined`));return}try{this.setData(f)}catch(v){l(v);return}(h=(m=z(this,cn).config).onSuccess)==null||h.call(m,f,this),(g=(d=z(this,cn).config).onSettled)==null||g.call(d,f,this.state.error,this),this.scheduleGc()},onError:l,onFail:(f,m)=>{gt(this,En,xr).call(this,{type:"failed",failureCount:f,error:m})},onPause:()=>{gt(this,En,xr).call(this,{type:"pause"})},onContinue:()=>{gt(this,En,xr).call(this,{type:"continue"})},retry:s.options.retry,retryDelay:s.options.retryDelay,networkMode:s.options.networkMode,canRun:()=>!0})),z(this,Pt).start()}},gi=new WeakMap,no=new WeakMap,cn=new WeakMap,ro=new WeakMap,Pt=new WeakMap,hu=new WeakMap,ao=new WeakMap,En=new WeakSet,xr=function(t){const n=r=>{switch(t.type){case"failed":return{...r,fetchFailureCount:t.failureCount,fetchFailureReason:t.error};case"pause":return{...r,fetchStatus:"paused"};case"continue":return{...r,fetchStatus:"fetching"};case"fetch":return{...r,...Koe(r.data,this.options),fetchMeta:t.meta??null};case"success":return ne(this,no,void 0),{...r,data:t.data,dataUpdateCount:r.dataUpdateCount+1,dataUpdatedAt:t.dataUpdatedAt??Date.now(),error:null,isInvalidated:!1,status:"success",...!t.manual&&{fetchStatus:"idle",fetchFailureCount:0,fetchFailureReason:null}};case"error":const a=t.error;return Wm(a)&&a.revert&&z(this,no)?{...z(this,no),fetchStatus:"idle"}:{...r,error:a,errorUpdateCount:r.errorUpdateCount+1,errorUpdatedAt:Date.now(),fetchFailureCount:r.fetchFailureCount+1,fetchFailureReason:a,fetchStatus:"idle",status:"error"};case"invalidate":return{...r,isInvalidated:!0};case"setState":return{...r,...t.state}}};this.state=n(this.state),Nt.batch(()=>{this.observers.forEach(r=>{r.onQueryUpdate()}),z(this,cn).notify({query:this,type:"updated",action:t})})},I2);function Koe(e,t){return{fetchFailureCount:0,fetchFailureReason:null,fetchStatus:YS(t.networkMode)?"fetching":"paused",...e===void 0&&{error:null,status:"pending"}}}function Xoe(e){const t=typeof e.initialData=="function"?e.initialData():e.initialData,n=t!==void 0,r=n?typeof e.initialDataUpdatedAt=="function"?e.initialDataUpdatedAt():e.initialDataUpdatedAt:0;return{data:t,dataUpdateCount:0,dataUpdatedAt:n?r??Date.now():0,error:null,errorUpdateCount:0,errorUpdatedAt:0,fetchFailureCount:0,fetchFailureReason:null,fetchMeta:null,isInvalidated:!1,status:n?"success":"pending",fetchStatus:"idle"}}var Qn,D2,Yoe=(D2=class extends Ad{constructor(t={}){super();me(this,Qn);this.config=t,ne(this,Qn,new Map)}build(t,n,r){const a=n.queryKey,o=n.queryHash??Ug(a,n);let i=this.get(o);return i||(i=new Voe({client:t,queryKey:a,queryHash:o,options:t.defaultQueryOptions(n),state:r,defaultOptions:t.getQueryDefaults(a)}),this.add(i)),i}add(t){z(this,Qn).has(t.queryHash)||(z(this,Qn).set(t.queryHash,t),this.notify({type:"added",query:t}))}remove(t){const n=z(this,Qn).get(t.queryHash);n&&(t.destroy(),n===t&&z(this,Qn).delete(t.queryHash),this.notify({type:"removed",query:t}))}clear(){Nt.batch(()=>{this.getAll().forEach(t=>{this.remove(t)})})}get(t){return z(this,Qn).get(t)}getAll(){return[...z(this,Qn).values()]}find(t){const n={exact:!0,...t};return this.getAll().find(r=>k2(n,r))}findAll(t={}){const n=this.getAll();return Object.keys(t).length>0?n.filter(r=>k2(t,r)):n}notify(t){Nt.batch(()=>{this.listeners.forEach(n=>{n(t)})})}onFocus(){Nt.batch(()=>{this.getAll().forEach(t=>{t.onFocus()})})}onOnline(){Nt.batch(()=>{this.getAll().forEach(t=>{t.onOnline()})})}},Qn=new WeakMap,D2),Zn,Tt,oo,Jn,Qr,L2,Qoe=(L2=class extends JS{constructor(t){super();me(this,Jn);me(this,Zn);me(this,Tt);me(this,oo);this.mutationId=t.mutationId,ne(this,Tt,t.mutationCache),ne(this,Zn,[]),this.state=t.state||Zoe(),this.setOptions(t.options),this.scheduleGc()}setOptions(t){this.options=t,this.updateGcTime(this.options.gcTime)}get meta(){return this.options.meta}addObserver(t){z(this,Zn).includes(t)||(z(this,Zn).push(t),this.clearGcTimeout(),z(this,Tt).notify({type:"observerAdded",mutation:this,observer:t}))}removeObserver(t){ne(this,Zn,z(this,Zn).filter(n=>n!==t)),this.scheduleGc(),z(this,Tt).notify({type:"observerRemoved",mutation:this,observer:t})}optionalRemove(){z(this,Zn).length||(this.state.status==="pending"?this.scheduleGc():z(this,Tt).remove(this))}continue(){var t;return((t=z(this,oo))==null?void 0:t.continue())??this.execute(this.state.variables)}async execute(t){var o,i,s,l,u,p,c,f,m,h,d,g,v,y,x,P,w,S,O,C;const n=()=>{gt(this,Jn,Qr).call(this,{type:"continue"})};ne(this,oo,ZS({fn:()=>this.options.mutationFn?this.options.mutationFn(t):Promise.reject(new Error("No mutationFn found")),onFail:(_,T)=>{gt(this,Jn,Qr).call(this,{type:"failed",failureCount:_,error:T})},onPause:()=>{gt(this,Jn,Qr).call(this,{type:"pause"})},onContinue:n,retry:this.options.retry??0,retryDelay:this.options.retryDelay,networkMode:this.options.networkMode,canRun:()=>z(this,Tt).canRun(this)}));const r=this.state.status==="pending",a=!z(this,oo).canStart();try{if(r)n();else{gt(this,Jn,Qr).call(this,{type:"pending",variables:t,isPaused:a}),await((i=(o=z(this,Tt).config).onMutate)==null?void 0:i.call(o,t,this));const T=await((l=(s=this.options).onMutate)==null?void 0:l.call(s,t));T!==this.state.context&&gt(this,Jn,Qr).call(this,{type:"pending",context:T,variables:t,isPaused:a})}const _=await z(this,oo).start();return await((p=(u=z(this,Tt).config).onSuccess)==null?void 0:p.call(u,_,t,this.state.context,this)),await((f=(c=this.options).onSuccess)==null?void 0:f.call(c,_,t,this.state.context)),await((h=(m=z(this,Tt).config).onSettled)==null?void 0:h.call(m,_,null,this.state.variables,this.state.context,this)),await((g=(d=this.options).onSettled)==null?void 0:g.call(d,_,null,t,this.state.context)),gt(this,Jn,Qr).call(this,{type:"success",data:_}),_}catch(_){try{throw await((y=(v=z(this,Tt).config).onError)==null?void 0:y.call(v,_,t,this.state.context,this)),await((P=(x=this.options).onError)==null?void 0:P.call(x,_,t,this.state.context)),await((S=(w=z(this,Tt).config).onSettled)==null?void 0:S.call(w,void 0,_,this.state.variables,this.state.context,this)),await((C=(O=this.options).onSettled)==null?void 0:C.call(O,void 0,_,t,this.state.context)),_}finally{gt(this,Jn,Qr).call(this,{type:"error",error:_})}}finally{z(this,Tt).runNext(this)}}},Zn=new WeakMap,Tt=new WeakMap,oo=new WeakMap,Jn=new WeakSet,Qr=function(t){const n=r=>{switch(t.type){case"failed":return{...r,failureCount:t.failureCount,failureReason:t.error};case"pause":return{...r,isPaused:!0};case"continue":return{...r,isPaused:!1};case"pending":return{...r,context:t.context,data:void 0,failureCount:0,failureReason:null,error:null,isPaused:t.isPaused,status:"pending",variables:t.variables,submittedAt:Date.now()};case"success":return{...r,data:t.data,failureCount:0,failureReason:null,error:null,status:"success",isPaused:!1};case"error":return{...r,data:void 0,error:t.error,failureCount:r.failureCount+1,failureReason:t.error,isPaused:!1,status:"error"}}};this.state=n(this.state),Nt.batch(()=>{z(this,Zn).forEach(r=>{r.onMutationUpdate(t)}),z(this,Tt).notify({mutation:this,type:"updated",action:t})})},L2);function Zoe(){return{context:void 0,data:void 0,error:null,failureCount:0,failureReason:null,isPaused:!1,status:"idle",variables:void 0,submittedAt:0}}var Pr,Tn,vu,F2,Joe=(F2=class extends Ad{constructor(t={}){super();me(this,Pr);me(this,Tn);me(this,vu);this.config=t,ne(this,Pr,new Set),ne(this,Tn,new Map),ne(this,vu,0)}build(t,n,r){const a=new Qoe({mutationCache:this,mutationId:++$u(this,vu)._,options:t.defaultMutationOptions(n),state:r});return this.add(a),a}add(t){z(this,Pr).add(t);const n=gc(t);if(typeof n=="string"){const r=z(this,Tn).get(n);r?r.push(t):z(this,Tn).set(n,[t])}this.notify({type:"added",mutation:t})}remove(t){if(z(this,Pr).delete(t)){const n=gc(t);if(typeof n=="string"){const r=z(this,Tn).get(n);if(r)if(r.length>1){const a=r.indexOf(t);a!==-1&&r.splice(a,1)}else r[0]===t&&z(this,Tn).delete(n)}}this.notify({type:"removed",mutation:t})}canRun(t){const n=gc(t);if(typeof n=="string"){const r=z(this,Tn).get(n),a=r==null?void 0:r.find(o=>o.state.status==="pending");return!a||a===t}else return!0}runNext(t){var r;const n=gc(t);if(typeof n=="string"){const a=(r=z(this,Tn).get(n))==null?void 0:r.find(o=>o!==t&&o.state.isPaused);return(a==null?void 0:a.continue())??Promise.resolve()}else return Promise.resolve()}clear(){Nt.batch(()=>{z(this,Pr).forEach(t=>{this.notify({type:"removed",mutation:t})}),z(this,Pr).clear(),z(this,Tn).clear()})}getAll(){return Array.from(z(this,Pr))}find(t){const n={exact:!0,...t};return this.getAll().find(r=>C2(n,r))}findAll(t={}){return this.getAll().filter(n=>C2(t,n))}notify(t){Nt.batch(()=>{this.listeners.forEach(n=>{n(t)})})}resumePausedMutations(){const t=this.getAll().filter(n=>n.state.isPaused);return Nt.batch(()=>Promise.all(t.map(n=>n.continue().catch(An))))}},Pr=new WeakMap,Tn=new WeakMap,vu=new WeakMap,F2);function gc(e){var t;return(t=e.options.scope)==null?void 0:t.id}function E2(e){return{onFetch:(t,n)=>{var p,c,f,m,h;const r=t.options,a=(f=(c=(p=t.fetchOptions)==null?void 0:p.meta)==null?void 0:c.fetchMore)==null?void 0:f.direction,o=((m=t.state.data)==null?void 0:m.pages)||[],i=((h=t.state.data)==null?void 0:h.pageParams)||[];let s={pages:[],pageParams:[]},l=0;const u=async()=>{let d=!1;const g=x=>{Object.defineProperty(x,"signal",{enumerable:!0,get:()=>(t.signal.aborted?d=!0:t.signal.addEventListener("abort",()=>{d=!0}),t.signal)})},v=KS(t.options,t.fetchOptions),y=async(x,P,w)=>{if(d)return Promise.reject();if(P==null&&x.pages.length)return Promise.resolve(x);const O=(()=>{const A={client:t.client,queryKey:t.queryKey,pageParam:P,direction:w?"backward":"forward",meta:t.options.meta};return g(A),A})(),C=await v(O),{maxPages:_}=t.options,T=w?Boe:Foe;return{pages:T(x.pages,C,_),pageParams:T(x.pageParams,P,_)}};if(a&&o.length){const x=a==="backward",P=x?eie:T2,w={pages:o,pageParams:i},S=P(r,w);s=await y(w,S,x)}else{const x=e??o.length;do{const P=l===0?i[0]??r.initialPageParam:T2(r,s);if(l>0&&P==null)break;s=await y(s,P),l++}while(l<x)}return s};t.options.persister?t.fetchFn=()=>{var d,g;return(g=(d=t.options).persister)==null?void 0:g.call(d,u,{client:t.client,queryKey:t.queryKey,meta:t.options.meta,signal:t.signal},n)}:t.fetchFn=u}}}function T2(e,{pages:t,pageParams:n}){const r=t.length-1;return t.length>0?e.getNextPageParam(t[r],t,n[r],n):void 0}function eie(e,{pages:t,pageParams:n}){var r;return t.length>0?(r=e.getPreviousPageParam)==null?void 0:r.call(e,t[0],t,n[0],n):void 0}var We,ia,sa,xi,wi,la,bi,Pi,B2,tie=(B2=class{constructor(e={}){me(this,We);me(this,ia);me(this,sa);me(this,xi);me(this,wi);me(this,la);me(this,bi);me(this,Pi);ne(this,We,e.queryCache||new Yoe),ne(this,ia,e.mutationCache||new Joe),ne(this,sa,e.defaultOptions||{}),ne(this,xi,new Map),ne(this,wi,new Map),ne(this,la,0)}mount(){$u(this,la)._++,z(this,la)===1&&(ne(this,bi,XS.subscribe(async e=>{e&&(await this.resumePausedMutations(),z(this,We).onFocus())})),ne(this,Pi,vf.subscribe(async e=>{e&&(await this.resumePausedMutations(),z(this,We).onOnline())})))}unmount(){var e,t;$u(this,la)._--,z(this,la)===0&&((e=z(this,bi))==null||e.call(this),ne(this,bi,void 0),(t=z(this,Pi))==null||t.call(this),ne(this,Pi,void 0))}isFetching(e){return z(this,We).findAll({...e,fetchStatus:"fetching"}).length}isMutating(e){return z(this,ia).findAll({...e,status:"pending"}).length}getQueryData(e){var n;const t=this.defaultQueryOptions({queryKey:e});return(n=z(this,We).get(t.queryHash))==null?void 0:n.state.data}ensureQueryData(e){const t=this.defaultQueryOptions(e),n=z(this,We).build(this,t),r=n.state.data;return r===void 0?this.fetchQuery(e):(e.revalidateIfStale&&n.isStaleByTime(Pv(t.staleTime,n))&&this.prefetchQuery(t),Promise.resolve(r))}getQueriesData(e){return z(this,We).findAll(e).map(({queryKey:t,state:n})=>{const r=n.data;return[t,r]})}setQueryData(e,t,n){const r=this.defaultQueryOptions({queryKey:e}),a=z(this,We).get(r.queryHash),o=a==null?void 0:a.state.data,i=$oe(t,o);if(i!==void 0)return z(this,We).build(this,r).setData(i,{...n,manual:!0})}setQueriesData(e,t,n){return Nt.batch(()=>z(this,We).findAll(e).map(({queryKey:r})=>[r,this.setQueryData(r,t,n)]))}getQueryState(e){var n;const t=this.defaultQueryOptions({queryKey:e});return(n=z(this,We).get(t.queryHash))==null?void 0:n.state}removeQueries(e){const t=z(this,We);Nt.batch(()=>{t.findAll(e).forEach(n=>{t.remove(n)})})}resetQueries(e,t){const n=z(this,We);return Nt.batch(()=>(n.findAll(e).forEach(r=>{r.reset()}),this.refetchQueries({type:"active",...e},t)))}cancelQueries(e,t={}){const n={revert:!0,...t},r=Nt.batch(()=>z(this,We).findAll(e).map(a=>a.cancel(n)));return Promise.all(r).then(An).catch(An)}invalidateQueries(e,t={}){return Nt.batch(()=>(z(this,We).findAll(e).forEach(n=>{n.invalidate()}),(e==null?void 0:e.refetchType)==="none"?Promise.resolve():this.refetchQueries({...e,type:(e==null?void 0:e.refetchType)??(e==null?void 0:e.type)??"active"},t)))}refetchQueries(e,t={}){const n={...t,cancelRefetch:t.cancelRefetch??!0},r=Nt.batch(()=>z(this,We).findAll(e).filter(a=>!a.isDisabled()&&!a.isStatic()).map(a=>{let o=a.fetch(void 0,n);return n.throwOnError||(o=o.catch(An)),a.state.fetchStatus==="paused"?Promise.resolve():o}));return Promise.all(r).then(An)}fetchQuery(e){const t=this.defaultQueryOptions(e);t.retry===void 0&&(t.retry=!1);const n=z(this,We).build(this,t);return n.isStaleByTime(Pv(t.staleTime,n))?n.fetch(t):Promise.resolve(n.state.data)}prefetchQuery(e){return this.fetchQuery(e).then(An).catch(An)}fetchInfiniteQuery(e){return e.behavior=E2(e.pages),this.fetchQuery(e)}prefetchInfiniteQuery(e){return this.fetchInfiniteQuery(e).then(An).catch(An)}ensureInfiniteQueryData(e){return e.behavior=E2(e.pages),this.ensureQueryData(e)}resumePausedMutations(){return vf.isOnline()?z(this,ia).resumePausedMutations():Promise.resolve()}getQueryCache(){return z(this,We)}getMutationCache(){return z(this,ia)}getDefaultOptions(){return z(this,sa)}setDefaultOptions(e){ne(this,sa,e)}setQueryDefaults(e,t){z(this,xi).set(du(e),{queryKey:e,defaultOptions:t})}getQueryDefaults(e){const t=[...z(this,xi).values()],n={};return t.forEach(r=>{mu(e,r.queryKey)&&Object.assign(n,r.defaultOptions)}),n}setMutationDefaults(e,t){z(this,wi).set(du(e),{mutationKey:e,defaultOptions:t})}getMutationDefaults(e){const t=[...z(this,wi).values()],n={};return t.forEach(r=>{mu(e,r.mutationKey)&&Object.assign(n,r.defaultOptions)}),n}defaultQueryOptions(e){if(e._defaulted)return e;const t={...z(this,sa).queries,...this.getQueryDefaults(e.queryKey),...e,_defaulted:!0};return t.queryHash||(t.queryHash=Ug(t.queryKey,t)),t.refetchOnReconnect===void 0&&(t.refetchOnReconnect=t.networkMode!=="always"),t.throwOnError===void 0&&(t.throwOnError=!!t.suspense),!t.networkMode&&t.persister&&(t.networkMode="offlineFirst"),t.queryFn===Wg&&(t.enabled=!1),t}defaultMutationOptions(e){return e!=null&&e._defaulted?e:{...z(this,sa).mutations,...(e==null?void 0:e.mutationKey)&&this.getMutationDefaults(e.mutationKey),...e,_defaulted:!0}}clear(){z(this,We).clear(),z(this,ia).clear()}},We=new WeakMap,ia=new WeakMap,sa=new WeakMap,xi=new WeakMap,wi=new WeakMap,la=new WeakMap,bi=new WeakMap,Pi=new WeakMap,B2),nie=k.createContext(void 0),rie=({client:e,children:t})=>(k.useEffect(()=>(e.mount(),()=>{e.unmount()}),[e]),b.jsx(nie.Provider,{value:e,children:t})),[Td,Aie]=is("Tooltip",[If]),qg=If(),eO="TooltipProvider",aie=700,j2="tooltip.open",[oie,tO]=Td(eO),nO=e=>{const{__scopeTooltip:t,delayDuration:n=aie,skipDelayDuration:r=300,disableHoverableContent:a=!1,children:o}=e,i=k.useRef(!0),s=k.useRef(!1),l=k.useRef(0);return k.useEffect(()=>{const u=l.current;return()=>window.clearTimeout(u)},[]),b.jsx(oie,{scope:t,isOpenDelayedRef:i,delayDuration:n,onOpen:k.useCallback(()=>{window.clearTimeout(l.current),i.current=!1},[]),onClose:k.useCallback(()=>{window.clearTimeout(l.current),l.current=window.setTimeout(()=>i.current=!0,r)},[r]),isPointerInTransitRef:s,onPointerInTransitChange:k.useCallback(u=>{s.current=u},[]),disableHoverableContent:a,children:o})};nO.displayName=eO;var rO="Tooltip",[Eie,jd]=Td(rO),Ov="TooltipTrigger",iie=k.forwardRef((e,t)=>{const{__scopeTooltip:n,...r}=e,a=jd(Ov,n),o=tO(Ov,n),i=qg(n),s=k.useRef(null),l=Xe(t,s,a.onTriggerChange),u=k.useRef(!1),p=k.useRef(!1),c=k.useCallback(()=>u.current=!1,[]);return k.useEffect(()=>()=>document.removeEventListener("pointerup",c),[c]),b.jsx(_6,{asChild:!0,...i,children:b.jsx(Ae.button,{"aria-describedby":a.open?a.contentId:void 0,"data-state":a.stateAttribute,...r,ref:l,onPointerMove:fe(e.onPointerMove,f=>{f.pointerType!=="touch"&&!p.current&&!o.isPointerInTransitRef.current&&(a.onTriggerEnter(),p.current=!0)}),onPointerLeave:fe(e.onPointerLeave,()=>{a.onTriggerLeave(),p.current=!1}),onPointerDown:fe(e.onPointerDown,()=>{a.open&&a.onClose(),u.current=!0,document.addEventListener("pointerup",c,{once:!0})}),onFocus:fe(e.onFocus,()=>{u.current||a.onOpen()}),onBlur:fe(e.onBlur,a.onClose),onClick:fe(e.onClick,a.onClose)})})});iie.displayName=Ov;var sie="TooltipPortal",[Tie,lie]=Td(sie,{forceMount:void 0}),ts="TooltipContent",aO=k.forwardRef((e,t)=>{const n=lie(ts,e.__scopeTooltip),{forceMount:r=n.forceMount,side:a="top",...o}=e,i=jd(ts,e.__scopeTooltip);return b.jsx(zg,{present:r||i.open,children:i.disableHoverableContent?b.jsx(oO,{side:a,...o,ref:t}):b.jsx(uie,{side:a,...o,ref:t})})}),uie=k.forwardRef((e,t)=>{const n=jd(ts,e.__scopeTooltip),r=tO(ts,e.__scopeTooltip),a=k.useRef(null),o=Xe(t,a),[i,s]=k.useState(null),{trigger:l,onClose:u}=n,p=a.current,{onPointerInTransitChange:c}=r,f=k.useCallback(()=>{s(null),c(!1)},[c]),m=k.useCallback((h,d)=>{const g=h.currentTarget,v={x:h.clientX,y:h.clientY},y=mie(v,g.getBoundingClientRect()),x=hie(v,y),P=vie(d.getBoundingClientRect()),w=gie([...x,...P]);s(w),c(!0)},[c]);return k.useEffect(()=>()=>f(),[f]),k.useEffect(()=>{if(l&&p){const h=g=>m(g,p),d=g=>m(g,l);return l.addEventListener("pointerleave",h),p.addEventListener("pointerleave",d),()=>{l.removeEventListener("pointerleave",h),p.removeEventListener("pointerleave",d)}}},[l,p,m,f]),k.useEffect(()=>{if(i){const h=d=>{const g=d.target,v={x:d.clientX,y:d.clientY},y=(l==null?void 0:l.contains(g))||(p==null?void 0:p.contains(g)),x=!yie(v,i);y?f():x&&(f(),u())};return document.addEventListener("pointermove",h),()=>document.removeEventListener("pointermove",h)}},[l,p,i,u,f]),b.jsx(oO,{...e,ref:o})}),[cie,pie]=Td(rO,{isInside:!1}),fie=BC("TooltipContent"),oO=k.forwardRef((e,t)=>{const{__scopeTooltip:n,children:r,"aria-label":a,onEscapeKeyDown:o,onPointerDownOutside:i,...s}=e,l=jd(ts,n),u=qg(n),{onClose:p}=l;return k.useEffect(()=>(document.addEventListener(j2,p),()=>document.removeEventListener(j2,p)),[p]),k.useEffect(()=>{if(l.trigger){const c=f=>{const m=f.target;m!=null&&m.contains(l.trigger)&&p()};return window.addEventListener("scroll",c,{capture:!0}),()=>window.removeEventListener("scroll",c,{capture:!0})}},[l.trigger,p]),b.jsx(wy,{asChild:!0,disableOutsidePointerEvents:!1,onEscapeKeyDown:o,onPointerDownOutside:i,onFocusOutside:c=>c.preventDefault(),onDismiss:p,children:b.jsxs(A6,{"data-state":l.stateAttribute,...u,...s,ref:t,style:{...s.style,"--radix-tooltip-content-transform-origin":"var(--radix-popper-transform-origin)","--radix-tooltip-content-available-width":"var(--radix-popper-available-width)","--radix-tooltip-content-available-height":"var(--radix-popper-available-height)","--radix-tooltip-trigger-width":"var(--radix-popper-anchor-width)","--radix-tooltip-trigger-height":"var(--radix-popper-anchor-height)"},children:[b.jsx(fie,{children:r}),b.jsx(cie,{scope:n,isInside:!0,children:b.jsx(CA,{id:l.contentId,role:"tooltip",children:a||r})})]})})});aO.displayName=ts;var iO="TooltipArrow",die=k.forwardRef((e,t)=>{const{__scopeTooltip:n,...r}=e,a=qg(n);return pie(iO,n).isInside?null:b.jsx(E6,{...a,...r,ref:t})});die.displayName=iO;function mie(e,t){const n=Math.abs(t.top-e.y),r=Math.abs(t.bottom-e.y),a=Math.abs(t.right-e.x),o=Math.abs(t.left-e.x);switch(Math.min(n,r,a,o)){case o:return"left";case a:return"right";case n:return"top";case r:return"bottom";default:throw new Error("unreachable")}}function hie(e,t,n=5){const r=[];switch(t){case"top":r.push({x:e.x-n,y:e.y+n},{x:e.x+n,y:e.y+n});break;case"bottom":r.push({x:e.x-n,y:e.y-n},{x:e.x+n,y:e.y-n});break;case"left":r.push({x:e.x+n,y:e.y-n},{x:e.x+n,y:e.y+n});break;case"right":r.push({x:e.x-n,y:e.y-n},{x:e.x-n,y:e.y+n});break}return r}function vie(e){const{top:t,right:n,bottom:r,left:a}=e;return[{x:a,y:t},{x:n,y:t},{x:n,y:r},{x:a,y:r}]}function yie(e,t){const{x:n,y:r}=e;let a=!1;for(let o=0,i=t.length-1;o<t.length;i=o++){const s=t[o],l=t[i],u=s.x,p=s.y,c=l.x,f=l.y;p>r!=f>r&&n<(c-u)*(r-p)/(f-p)+u&&(a=!a)}return a}function gie(e){const t=e.slice();return t.sort((n,r)=>n.x<r.x?-1:n.x>r.x?1:n.y<r.y?-1:n.y>r.y?1:0),xie(t)}function xie(e){if(e.length<=1)return e.slice();const t=[];for(let r=0;r<e.length;r++){const a=e[r];for(;t.length>=2;){const o=t[t.length-1],i=t[t.length-2];if((o.x-i.x)*(a.y-i.y)>=(o.y-i.y)*(a.x-i.x))t.pop();else break}t.push(a)}t.pop();const n=[];for(let r=e.length-1;r>=0;r--){const a=e[r];for(;n.length>=2;){const o=n[n.length-1],i=n[n.length-2];if((o.x-i.x)*(a.y-i.y)>=(o.y-i.y)*(a.x-i.x))n.pop();else break}n.push(a)}return n.pop(),t.length===1&&n.length===1&&t[0].x===n[0].x&&t[0].y===n[0].y?t:t.concat(n)}var wie=nO,sO=aO;const bie=wie,Pie=k.forwardRef(({className:e,sideOffset:t=4,...n},r)=>b.jsx(sO,{ref:r,sideOffset:t,className:Se("z-50 overflow-hidden rounded-md border bg-popover px-3 py-1.5 text-sm text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",e),...n}));Pie.displayName=sO.displayName;const Sie=e=>{const t=k.useRef(null),n=k.useRef(null),r=k.useRef(null);return k.useEffect(()=>{if(t.current){try{const a=t.current.attachShadow({mode:"open"});n.current=a;const o=document.createElement("div");o.id="shadow-root",o.className="benchmark-dashboard",a.appendChild(o),qS.getAllCSS().then(i=>{const s=document.createElement("style");s.textContent=i,a.insertBefore(s,o);const l=Ic.createRoot(o);r.current=l,l.render(b.jsx(E.StrictMode,{children:b.jsx(rie,{client:new tie,children:b.jsx(bie,{children:b.jsx(Noe,{...e})})})}))}).catch(i=>{console.error("Failed to inject CSS into Shadow DOM:",i)})}catch(a){console.error("Failed to create Shadow DOM:",a)}return()=>{r.current&&r.current.unmount()}}},[]),b.jsx("div",{ref:t,style:{width:"100%",minHeight:"400px"}})},Oie=`*, ::before, ::after {
  --tw-border-spacing-x: 0;
  --tw-border-spacing-y: 0;
  --tw-translate-x: 0;
  --tw-translate-y: 0;
  --tw-rotate: 0;
  --tw-skew-x: 0;
  --tw-skew-y: 0;
  --tw-scale-x: 1;
  --tw-scale-y: 1;
  --tw-pan-x:  ;
  --tw-pan-y:  ;
  --tw-pinch-zoom:  ;
  --tw-scroll-snap-strictness: proximity;
  --tw-gradient-from-position:  ;
  --tw-gradient-via-position:  ;
  --tw-gradient-to-position:  ;
  --tw-ordinal:  ;
  --tw-slashed-zero:  ;
  --tw-numeric-figure:  ;
  --tw-numeric-spacing:  ;
  --tw-numeric-fraction:  ;
  --tw-ring-inset:  ;
  --tw-ring-offset-width: 0px;
  --tw-ring-offset-color: #fff;
  --tw-ring-color: rgb(59 130 246 / 0.5);
  --tw-ring-offset-shadow: 0 0 #0000;
  --tw-ring-shadow: 0 0 #0000;
  --tw-shadow: 0 0 #0000;
  --tw-shadow-colored: 0 0 #0000;
  --tw-blur:  ;
  --tw-brightness:  ;
  --tw-contrast:  ;
  --tw-grayscale:  ;
  --tw-hue-rotate:  ;
  --tw-invert:  ;
  --tw-saturate:  ;
  --tw-sepia:  ;
  --tw-drop-shadow:  ;
  --tw-backdrop-blur:  ;
  --tw-backdrop-brightness:  ;
  --tw-backdrop-contrast:  ;
  --tw-backdrop-grayscale:  ;
  --tw-backdrop-hue-rotate:  ;
  --tw-backdrop-invert:  ;
  --tw-backdrop-opacity:  ;
  --tw-backdrop-saturate:  ;
  --tw-backdrop-sepia:  ;
  --tw-contain-size:  ;
  --tw-contain-layout:  ;
  --tw-contain-paint:  ;
  --tw-contain-style:  ;
}

::backdrop {
  --tw-border-spacing-x: 0;
  --tw-border-spacing-y: 0;
  --tw-translate-x: 0;
  --tw-translate-y: 0;
  --tw-rotate: 0;
  --tw-skew-x: 0;
  --tw-skew-y: 0;
  --tw-scale-x: 1;
  --tw-scale-y: 1;
  --tw-pan-x:  ;
  --tw-pan-y:  ;
  --tw-pinch-zoom:  ;
  --tw-scroll-snap-strictness: proximity;
  --tw-gradient-from-position:  ;
  --tw-gradient-via-position:  ;
  --tw-gradient-to-position:  ;
  --tw-ordinal:  ;
  --tw-slashed-zero:  ;
  --tw-numeric-figure:  ;
  --tw-numeric-spacing:  ;
  --tw-numeric-fraction:  ;
  --tw-ring-inset:  ;
  --tw-ring-offset-width: 0px;
  --tw-ring-offset-color: #fff;
  --tw-ring-color: rgb(59 130 246 / 0.5);
  --tw-ring-offset-shadow: 0 0 #0000;
  --tw-ring-shadow: 0 0 #0000;
  --tw-shadow: 0 0 #0000;
  --tw-shadow-colored: 0 0 #0000;
  --tw-blur:  ;
  --tw-brightness:  ;
  --tw-contrast:  ;
  --tw-grayscale:  ;
  --tw-hue-rotate:  ;
  --tw-invert:  ;
  --tw-saturate:  ;
  --tw-sepia:  ;
  --tw-drop-shadow:  ;
  --tw-backdrop-blur:  ;
  --tw-backdrop-brightness:  ;
  --tw-backdrop-contrast:  ;
  --tw-backdrop-grayscale:  ;
  --tw-backdrop-hue-rotate:  ;
  --tw-backdrop-invert:  ;
  --tw-backdrop-opacity:  ;
  --tw-backdrop-saturate:  ;
  --tw-backdrop-sepia:  ;
  --tw-contain-size:  ;
  --tw-contain-layout:  ;
  --tw-contain-paint:  ;
  --tw-contain-style:  ;
}

:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --card: 0 0% 100%;
  --card-foreground: 222.2 84% 4.9%;
  --popover: 0 0% 100%;
  --popover-foreground: 222.2 84% 4.9%;
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  --secondary: 210 40% 96.1%;
  --secondary-foreground: 222.2 47.4% 11.2%;
  --muted: 210 40% 96.1%;
  --muted-foreground: 215.4 16.3% 46.9%;
  --accent: 210 40% 96.1%;
  --accent-foreground: 222.2 47.4% 11.2%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  --border: 214.3 31.8% 91.4%;
  --input: 214.3 31.8% 91.4%;
  --ring: 222.2 84% 4.9%;
  --radius: 0.5rem;
  --sidebar-background: 0 0% 98%;
  --sidebar-foreground: 240 5.3% 26.1%;
  --sidebar-primary: 240 5.9% 10%;
  --sidebar-primary-foreground: 0 0% 98%;
  --sidebar-accent: 240 4.8% 95.9%;
  --sidebar-accent-foreground: 240 5.9% 10%;
  --sidebar-border: 220 13% 91%;
  --sidebar-ring: 217.2 91.2% 59.8%;
}

body {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --card: 0 0% 100%;
  --card-foreground: 222.2 84% 4.9%;
  --popover: 0 0% 100%;
  --popover-foreground: 222.2 84% 4.9%;
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  --secondary: 210 40% 96.1%;
  --secondary-foreground: 222.2 47.4% 11.2%;
  --muted: 210 40% 96.1%;
  --muted-foreground: 215.4 16.3% 46.9%;
  --accent: 210 40% 96.1%;
  --accent-foreground: 222.2 47.4% 11.2%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  --border: 214.3 31.8% 91.4%;
  --input: 214.3 31.8% 91.4%;
  --ring: 222.2 84% 4.9%;
  --radius: 0.5rem;
  --sidebar-background: 0 0% 98%;
  --sidebar-foreground: 240 5.3% 26.1%;
  --sidebar-primary: 240 5.9% 10%;
  --sidebar-primary-foreground: 0 0% 98%;
  --sidebar-accent: 240 4.8% 95.9%;
  --sidebar-accent-foreground: 240 5.9% 10%;
  --sidebar-border: 220 13% 91%;
  --sidebar-ring: 217.2 91.2% 59.8%;
}

.dark, body.dark {
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  --card: 222.2 84% 4.9%;
  --card-foreground: 210 40% 98%;
  --popover: 222.2 84% 4.9%;
  --popover-foreground: 210 40% 98%;
  --primary: 210 40% 98%;
  --primary-foreground: 222.2 47.4% 11.2%;
  --secondary: 217.2 32.6% 17.5%;
  --secondary-foreground: 210 40% 98%;
  --muted: 217.2 32.6% 17.5%;
  --muted-foreground: 215 20.2% 65.1%;
  --accent: 217.2 32.6% 17.5%;
  --accent-foreground: 210 40% 98%;
  --destructive: 0 62.8% 30.6%;
  --destructive-foreground: 210 40% 98%;
  --border: 217.2 32.6% 17.5%;
  --input: 217.2 32.6% 17.5%;
  --ring: 212.7 26.8% 83.9%;
  --sidebar-background: 240 5.9% 10%;
  --sidebar-foreground: 240 4.8% 95.9%;
  --sidebar-primary: 224.3 76.3% 48%;
  --sidebar-primary-foreground: 0 0% 100%;
  --sidebar-accent: 240 3.7% 15.9%;
  --sidebar-accent-foreground: 240 4.8% 95.9%;
  --sidebar-border: 240 3.7% 15.9%;
  --sidebar-ring: 217.2 91.2% 59.8%;
}

* {
  border-color: hsl(var(--border));
}

body {
  background-color: hsl(var(--background));
  color: hsl(var(--foreground));
}

.container {
  width: 100%;
  margin-right: auto;
  margin-left: auto;
  padding-right: 2rem;
  padding-left: 2rem;
}

@media (min-width: 1400px) {
  .container {
    max-width: 1400px;
  }
}

/* Radix popper content wrapper styling - these elements are rendered via React portals */

[data-radix-popper-content-wrapper] {
  /* Ensure high z-index for proper layering */
  z-index: 9999 !important;
  /* Apply CSS variables that should be inherited from the document */
  --popover: 0 0% 100%;
  --popover-foreground: 222.2 84% 4.9%;
  --border: 214.3 31.8% 91.4%;
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --muted: 210 40% 96.1%;
  --muted-foreground: 215.4 16.3% 46.9%;
  --accent: 210 40% 96.1%;
  --accent-foreground: 222.2 47.4% 11.2%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 210 40% 98%;
  --primary: 222.2 47.4% 11.2%;
  --primary-foreground: 210 40% 98%;
  --secondary: 210 40% 96.1%;
  --secondary-foreground: 222.2 47.4% 11.2%;
  --ring: 222.2 84% 4.9%;
  --radius: 0.5rem;
}

/* Dark mode support for portals */

.dark [data-radix-popper-content-wrapper],
  body.dark [data-radix-popper-content-wrapper] {
  --popover: 222.2 84% 4.9%;
  --popover-foreground: 210 40% 98%;
  --border: 217.2 32.6% 17.5%;
  --background: 222.2 84% 4.9%;
  --foreground: 210 40% 98%;
  --muted: 217.2 32.6% 17.5%;
  --muted-foreground: 215 20.2% 65.1%;
  --accent: 217.2 32.6% 17.5%;
  --accent-foreground: 210 40% 98%;
  --destructive: 0 62.8% 30.6%;
  --destructive-foreground: 210 40% 98%;
  --primary: 210 40% 98%;
  --primary-foreground: 222.2 47.4% 11.2%;
  --secondary: 217.2 32.6% 17.5%;
  --secondary-foreground: 210 40% 98%;
  --ring: 212.7 26.8% 83.9%;
}

/* Ensure portal content gets proper styling */

[data-radix-popper-content-wrapper] .bg-popover {
  background-color: hsl(var(--popover)) !important;
}

[data-radix-popper-content-wrapper] .text-popover-foreground {
  color: hsl(var(--popover-foreground)) !important;
}

[data-radix-popper-content-wrapper] .border {
  border-color: hsl(var(--border)) !important;
}

[data-radix-popper-content-wrapper] .shadow-md {
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1) !important;
}

[data-radix-popper-content-wrapper] .rounded-md {
  border-radius: calc(var(--radius) - 2px) !important;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border-width: 0;
}

.not-sr-only {
  position: static;
  width: auto;
  height: auto;
  padding: 0;
  margin: 0;
  overflow: visible;
  clip: auto;
  white-space: normal;
}

.pointer-events-none {
  pointer-events: none;
}

.pointer-events-auto {
  pointer-events: auto;
}

.visible {
  visibility: visible;
}

.invisible {
  visibility: hidden;
}

.static {
  position: static;
}

.fixed {
  position: fixed;
}

.absolute {
  position: absolute;
}

.relative {
  position: relative;
}

.sticky {
  position: sticky;
}

.inset-0 {
  inset: 0px;
}

.inset-x-0 {
  left: 0px;
  right: 0px;
}

.inset-y-0 {
  top: 0px;
  bottom: 0px;
}

.-bottom-12 {
  bottom: -3rem;
}

.-left-12 {
  left: -3rem;
}

.-right-12 {
  right: -3rem;
}

.-top-12 {
  top: -3rem;
}

.bottom-0 {
  bottom: 0px;
}

.bottom-8 {
  bottom: 2rem;
}

.left-0 {
  left: 0px;
}

.left-1 {
  left: 0.25rem;
}

.left-1\\/2 {
  left: 50%;
}

.left-2 {
  left: 0.5rem;
}

.left-4 {
  left: 1rem;
}

.left-\\[50\\%\\] {
  left: 50%;
}

.right-0 {
  right: 0px;
}

.right-1 {
  right: 0.25rem;
}

.right-2 {
  right: 0.5rem;
}

.right-3 {
  right: 0.75rem;
}

.right-4 {
  right: 1rem;
}

.top-0 {
  top: 0px;
}

.top-1\\.5 {
  top: 0.375rem;
}

.top-1\\/2 {
  top: 50%;
}

.top-2 {
  top: 0.5rem;
}

.top-3\\.5 {
  top: 0.875rem;
}

.top-4 {
  top: 1rem;
}

.top-\\[1px\\] {
  top: 1px;
}

.top-\\[50\\%\\] {
  top: 50%;
}

.top-\\[60\\%\\] {
  top: 60%;
}

.top-full {
  top: 100%;
}

.z-10 {
  z-index: 10;
}

.z-20 {
  z-index: 20;
}

.z-50 {
  z-index: 50;
}

.z-\\[100\\] {
  z-index: 100;
}

.z-\\[1\\] {
  z-index: 1;
}

.m-0 {
  margin: 0px;
}

.m-1 {
  margin: 0.25rem;
}

.m-2 {
  margin: 0.5rem;
}

.m-4 {
  margin: 1rem;
}

.m-6 {
  margin: 1.5rem;
}

.m-8 {
  margin: 2rem;
}

.-mx-1 {
  margin-left: -0.25rem;
  margin-right: -0.25rem;
}

.mx-1 {
  margin-left: 0.25rem;
  margin-right: 0.25rem;
}

.mx-2 {
  margin-left: 0.5rem;
  margin-right: 0.5rem;
}

.mx-3\\.5 {
  margin-left: 0.875rem;
  margin-right: 0.875rem;
}

.mx-auto {
  margin-left: auto;
  margin-right: auto;
}

.my-0\\.5 {
  margin-top: 0.125rem;
  margin-bottom: 0.125rem;
}

.my-1 {
  margin-top: 0.25rem;
  margin-bottom: 0.25rem;
}

.-ml-4 {
  margin-left: -1rem;
}

.-mt-4 {
  margin-top: -1rem;
}

.mb-1 {
  margin-bottom: 0.25rem;
}

.mb-10 {
  margin-bottom: 2.5rem;
}

.mb-12 {
  margin-bottom: 3rem;
}

.mb-2 {
  margin-bottom: 0.5rem;
}

.mb-3 {
  margin-bottom: 0.75rem;
}

.mb-4 {
  margin-bottom: 1rem;
}

.mb-5 {
  margin-bottom: 1.25rem;
}

.mb-6 {
  margin-bottom: 1.5rem;
}

.mb-8 {
  margin-bottom: 2rem;
}

.ml-1 {
  margin-left: 0.25rem;
}

.ml-3 {
  margin-left: 0.75rem;
}

.ml-6 {
  margin-left: 1.5rem;
}

.ml-auto {
  margin-left: auto;
}

.mr-0 {
  margin-right: 0px;
}

.mr-2 {
  margin-right: 0.5rem;
}

.mr-3 {
  margin-right: 0.75rem;
}

.mr-4 {
  margin-right: 1rem;
}

.mt-0\\.5 {
  margin-top: 0.125rem;
}

.mt-1 {
  margin-top: 0.25rem;
}

.mt-1\\.5 {
  margin-top: 0.375rem;
}

.mt-2 {
  margin-top: 0.5rem;
}

.mt-24 {
  margin-top: 6rem;
}

.mt-3 {
  margin-top: 0.75rem;
}

.mt-4 {
  margin-top: 1rem;
}

.mt-5 {
  margin-top: 1.25rem;
}

.mt-6 {
  margin-top: 1.5rem;
}

.mt-8 {
  margin-top: 2rem;
}

.mt-auto {
  margin-top: auto;
}

.block {
  display: block;
}

.inline-block {
  display: inline-block;
}

.inline {
  display: inline;
}

.flex {
  display: flex;
}

.inline-flex {
  display: inline-flex;
}

.table {
  display: table;
}

.grid {
  display: grid;
}

.hidden {
  display: none;
}

.aspect-square {
  aspect-ratio: 1 / 1;
}

.aspect-video {
  aspect-ratio: 16 / 9;
}

.size-4 {
  width: 1rem;
  height: 1rem;
}

.h-1\\.5 {
  height: 0.375rem;
}

.h-10 {
  height: 2.5rem;
}

.h-11 {
  height: 2.75rem;
}

.h-12 {
  height: 3rem;
}

.h-2 {
  height: 0.5rem;
}

.h-2\\.5 {
  height: 0.625rem;
}

.h-3 {
  height: 0.75rem;
}

.h-3\\.5 {
  height: 0.875rem;
}

.h-4 {
  height: 1rem;
}

.h-5 {
  height: 1.25rem;
}

.h-6 {
  height: 1.5rem;
}

.h-7 {
  height: 1.75rem;
}

.h-8 {
  height: 2rem;
}

.h-9 {
  height: 2.25rem;
}

.h-\\[1px\\] {
  height: 1px;
}

.h-\\[var\\(--radix-navigation-menu-viewport-height\\)\\] {
  height: var(--radix-navigation-menu-viewport-height);
}

.h-\\[var\\(--radix-select-trigger-height\\)\\] {
  height: var(--radix-select-trigger-height);
}

.h-auto {
  height: auto;
}

.h-full {
  height: 100%;
}

.h-px {
  height: 1px;
}

.h-svh {
  height: 100svh;
}

.max-h-96 {
  max-height: 24rem;
}

.max-h-\\[300px\\] {
  max-height: 300px;
}

.max-h-screen {
  max-height: 100vh;
}

.min-h-0 {
  min-height: 0px;
}

.min-h-\\[400px\\] {
  min-height: 400px;
}

.min-h-\\[80px\\] {
  min-height: 80px;
}

.min-h-full {
  min-height: 100%;
}

.min-h-svh {
  min-height: 100svh;
}

.w-0 {
  width: 0px;
}

.w-1 {
  width: 0.25rem;
}

.w-10 {
  width: 2.5rem;
}

.w-11 {
  width: 2.75rem;
}

.w-12 {
  width: 3rem;
}

.w-2 {
  width: 0.5rem;
}

.w-2\\.5 {
  width: 0.625rem;
}

.w-3 {
  width: 0.75rem;
}

.w-3\\.5 {
  width: 0.875rem;
}

.w-3\\/4 {
  width: 75%;
}

.w-32 {
  width: 8rem;
}

.w-4 {
  width: 1rem;
}

.w-5 {
  width: 1.25rem;
}

.w-6 {
  width: 1.5rem;
}

.w-64 {
  width: 16rem;
}

.w-7 {
  width: 1.75rem;
}

.w-72 {
  width: 18rem;
}

.w-8 {
  width: 2rem;
}

.w-9 {
  width: 2.25rem;
}

.w-\\[--sidebar-width\\] {
  width: var(--sidebar-width);
}

.w-\\[100px\\] {
  width: 100px;
}

.w-\\[1px\\] {
  width: 1px;
}

.w-auto {
  width: auto;
}

.w-fit {
  width: -moz-fit-content;
  width: fit-content;
}

.w-full {
  width: 100%;
}

.w-max {
  width: -moz-max-content;
  width: max-content;
}

.w-px {
  width: 1px;
}

.min-w-0 {
  min-width: 0px;
}

.min-w-5 {
  min-width: 1.25rem;
}

.min-w-\\[12rem\\] {
  min-width: 12rem;
}

.min-w-\\[8rem\\] {
  min-width: 8rem;
}

.min-w-\\[var\\(--radix-select-trigger-width\\)\\] {
  min-width: var(--radix-select-trigger-width);
}

.max-w-2xl {
  max-width: 42rem;
}

.max-w-3xl {
  max-width: 48rem;
}

.max-w-4xl {
  max-width: 56rem;
}

.max-w-7xl {
  max-width: 80rem;
}

.max-w-\\[--skeleton-width\\] {
  max-width: var(--skeleton-width);
}

.max-w-lg {
  max-width: 32rem;
}

.max-w-max {
  max-width: -moz-max-content;
  max-width: max-content;
}

.max-w-none {
  max-width: none;
}

.max-w-sm {
  max-width: 24rem;
}

.max-w-xs {
  max-width: 20rem;
}

.flex-1 {
  flex: 1 1 0%;
}

.flex-shrink-0 {
  flex-shrink: 0;
}

.shrink-0 {
  flex-shrink: 0;
}

.flex-grow {
  flex-grow: 1;
}

.grow {
  flex-grow: 1;
}

.grow-0 {
  flex-grow: 0;
}

.basis-full {
  flex-basis: 100%;
}

.caption-bottom {
  caption-side: bottom;
}

.border-collapse {
  border-collapse: collapse;
}

.-translate-x-1\\/2 {
  --tw-translate-x: -50%;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.-translate-x-px {
  --tw-translate-x: -1px;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.-translate-y-1\\/2 {
  --tw-translate-y: -50%;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.translate-x-\\[-50\\%\\] {
  --tw-translate-x: -50%;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.translate-x-px {
  --tw-translate-x: 1px;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.translate-y-\\[-50\\%\\] {
  --tw-translate-y: -50%;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.rotate-45 {
  --tw-rotate: 45deg;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.rotate-90 {
  --tw-rotate: 90deg;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.transform {
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

@keyframes pulse {
  50% {
    opacity: .5;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}

.cursor-default {
  cursor: default;
}

.cursor-pointer {
  cursor: pointer;
}

.touch-none {
  touch-action: none;
}

.select-none {
  -webkit-user-select: none;
     -moz-user-select: none;
          user-select: none;
}

.list-none {
  list-style-type: none;
}

.grid-cols-1 {
  grid-template-columns: repeat(1, minmax(0, 1fr));
}

.grid-cols-2 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.grid-cols-3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.flex-row {
  flex-direction: row;
}

.flex-col {
  flex-direction: column;
}

.flex-col-reverse {
  flex-direction: column-reverse;
}

.flex-wrap {
  flex-wrap: wrap;
}

.items-start {
  align-items: flex-start;
}

.items-end {
  align-items: flex-end;
}

.items-center {
  align-items: center;
}

.items-stretch {
  align-items: stretch;
}

.justify-start {
  justify-content: flex-start;
}

.justify-center {
  justify-content: center;
}

.justify-between {
  justify-content: space-between;
}

.gap-1 {
  gap: 0.25rem;
}

.gap-1\\.5 {
  gap: 0.375rem;
}

.gap-2 {
  gap: 0.5rem;
}

.gap-3 {
  gap: 0.75rem;
}

.gap-4 {
  gap: 1rem;
}

.gap-6 {
  gap: 1.5rem;
}

.gap-8 {
  gap: 2rem;
}

.space-x-1 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-x-reverse: 0;
  margin-right: calc(0.25rem * var(--tw-space-x-reverse));
  margin-left: calc(0.25rem * calc(1 - var(--tw-space-x-reverse)));
}

.space-x-2 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-x-reverse: 0;
  margin-right: calc(0.5rem * var(--tw-space-x-reverse));
  margin-left: calc(0.5rem * calc(1 - var(--tw-space-x-reverse)));
}

.space-x-3 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-x-reverse: 0;
  margin-right: calc(0.75rem * var(--tw-space-x-reverse));
  margin-left: calc(0.75rem * calc(1 - var(--tw-space-x-reverse)));
}

.space-x-4 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-x-reverse: 0;
  margin-right: calc(1rem * var(--tw-space-x-reverse));
  margin-left: calc(1rem * calc(1 - var(--tw-space-x-reverse)));
}

.space-y-0 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(0px * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(0px * var(--tw-space-y-reverse));
}

.space-y-1 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(0.25rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(0.25rem * var(--tw-space-y-reverse));
}

.space-y-1\\.5 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(0.375rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(0.375rem * var(--tw-space-y-reverse));
}

.space-y-2 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(0.5rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(0.5rem * var(--tw-space-y-reverse));
}

.space-y-3 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(0.75rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(0.75rem * var(--tw-space-y-reverse));
}

.space-y-4 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(1rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(1rem * var(--tw-space-y-reverse));
}

.space-y-6 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(1.5rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(1.5rem * var(--tw-space-y-reverse));
}

.space-y-8 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-y-reverse: 0;
  margin-top: calc(2rem * calc(1 - var(--tw-space-y-reverse)));
  margin-bottom: calc(2rem * var(--tw-space-y-reverse));
}

.self-start {
  align-self: flex-start;
}

.overflow-auto {
  overflow: auto;
}

.overflow-hidden {
  overflow: hidden;
}

.overflow-x-auto {
  overflow-x: auto;
}

.overflow-y-auto {
  overflow-y: auto;
}

.overflow-x-hidden {
  overflow-x: hidden;
}

.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.whitespace-nowrap {
  white-space: nowrap;
}

.whitespace-pre-wrap {
  white-space: pre-wrap;
}

.break-words {
  overflow-wrap: break-word;
}

.rounded {
  border-radius: 0.25rem;
}

.rounded-2xl {
  border-radius: 1rem;
}

.rounded-\\[2px\\] {
  border-radius: 2px;
}

.rounded-\\[inherit\\] {
  border-radius: inherit;
}

.rounded-full {
  border-radius: 9999px;
}

.rounded-lg {
  border-radius: var(--radius);
}

.rounded-md {
  border-radius: calc(var(--radius) - 2px);
}

.rounded-none {
  border-radius: 0px;
}

.rounded-sm {
  border-radius: calc(var(--radius) - 4px);
}

.rounded-xl {
  border-radius: 0.75rem;
}

.rounded-t-\\[10px\\] {
  border-top-left-radius: 10px;
  border-top-right-radius: 10px;
}

.rounded-t-lg {
  border-top-left-radius: var(--radius);
  border-top-right-radius: var(--radius);
}

.rounded-tl-sm {
  border-top-left-radius: calc(var(--radius) - 4px);
}

.border {
  border-width: 1px;
}

.border-2 {
  border-width: 2px;
}

.border-\\[1\\.5px\\] {
  border-width: 1.5px;
}

.border-y {
  border-top-width: 1px;
  border-bottom-width: 1px;
}

.border-b {
  border-bottom-width: 1px;
}

.border-b-2 {
  border-bottom-width: 2px;
}

.border-l {
  border-left-width: 1px;
}

.border-r {
  border-right-width: 1px;
}

.border-t {
  border-top-width: 1px;
}

.border-dashed {
  border-style: dashed;
}

.border-\\[--color-border\\] {
  border-color: var(--color-border);
}

.border-blue-100 {
  --tw-border-opacity: 1;
  border-color: rgb(219 234 254 / var(--tw-border-opacity, 1));
}

.border-blue-200 {
  --tw-border-opacity: 1;
  border-color: rgb(191 219 254 / var(--tw-border-opacity, 1));
}

.border-blue-300 {
  --tw-border-opacity: 1;
  border-color: rgb(147 197 253 / var(--tw-border-opacity, 1));
}

.border-blue-600 {
  --tw-border-opacity: 1;
  border-color: rgb(37 99 235 / var(--tw-border-opacity, 1));
}

.border-border\\/50 {
  border-color: hsl(var(--border) / 0.5);
}

.border-destructive {
  border-color: hsl(var(--destructive));
}

.border-destructive\\/50 {
  border-color: hsl(var(--destructive) / 0.5);
}

.border-gray-100 {
  --tw-border-opacity: 1;
  border-color: rgb(243 244 246 / var(--tw-border-opacity, 1));
}

.border-gray-200 {
  --tw-border-opacity: 1;
  border-color: rgb(229 231 235 / var(--tw-border-opacity, 1));
}

.border-gray-200\\/30 {
  border-color: rgb(229 231 235 / 0.3);
}

.border-gray-300 {
  --tw-border-opacity: 1;
  border-color: rgb(209 213 219 / var(--tw-border-opacity, 1));
}

.border-gray-700 {
  --tw-border-opacity: 1;
  border-color: rgb(55 65 81 / var(--tw-border-opacity, 1));
}

.border-green-100 {
  --tw-border-opacity: 1;
  border-color: rgb(220 252 231 / var(--tw-border-opacity, 1));
}

.border-green-200 {
  --tw-border-opacity: 1;
  border-color: rgb(187 247 208 / var(--tw-border-opacity, 1));
}

.border-green-300 {
  --tw-border-opacity: 1;
  border-color: rgb(134 239 172 / var(--tw-border-opacity, 1));
}

.border-input {
  border-color: hsl(var(--input));
}

.border-orange-200 {
  --tw-border-opacity: 1;
  border-color: rgb(254 215 170 / var(--tw-border-opacity, 1));
}

.border-primary {
  border-color: hsl(var(--primary));
}

.border-purple-200 {
  --tw-border-opacity: 1;
  border-color: rgb(233 213 255 / var(--tw-border-opacity, 1));
}

.border-purple-600 {
  --tw-border-opacity: 1;
  border-color: rgb(147 51 234 / var(--tw-border-opacity, 1));
}

.border-red-200 {
  --tw-border-opacity: 1;
  border-color: rgb(254 202 202 / var(--tw-border-opacity, 1));
}

.border-red-300 {
  --tw-border-opacity: 1;
  border-color: rgb(252 165 165 / var(--tw-border-opacity, 1));
}

.border-sidebar-border {
  border-color: hsl(var(--sidebar-border));
}

.border-teal-200 {
  --tw-border-opacity: 1;
  border-color: rgb(153 246 228 / var(--tw-border-opacity, 1));
}

.border-transparent {
  border-color: transparent;
}

.border-l-transparent {
  border-left-color: transparent;
}

.border-t-transparent {
  border-top-color: transparent;
}

.bg-\\[--color-bg\\] {
  background-color: var(--color-bg);
}

.bg-accent {
  background-color: hsl(var(--accent));
}

.bg-background {
  background-color: hsl(var(--background));
}

.bg-black\\/80 {
  background-color: rgb(0 0 0 / 0.8);
}

.bg-blue-100 {
  --tw-bg-opacity: 1;
  background-color: rgb(219 234 254 / var(--tw-bg-opacity, 1));
}

.bg-blue-50 {
  --tw-bg-opacity: 1;
  background-color: rgb(239 246 255 / var(--tw-bg-opacity, 1));
}

.bg-blue-50\\/50 {
  background-color: rgb(239 246 255 / 0.5);
}

.bg-blue-600 {
  --tw-bg-opacity: 1;
  background-color: rgb(37 99 235 / var(--tw-bg-opacity, 1));
}

.bg-border {
  background-color: hsl(var(--border));
}

.bg-card {
  background-color: hsl(var(--card));
}

.bg-destructive {
  background-color: hsl(var(--destructive));
}

.bg-foreground {
  background-color: hsl(var(--foreground));
}

.bg-gray-100 {
  --tw-bg-opacity: 1;
  background-color: rgb(243 244 246 / var(--tw-bg-opacity, 1));
}

.bg-gray-50 {
  --tw-bg-opacity: 1;
  background-color: rgb(249 250 251 / var(--tw-bg-opacity, 1));
}

.bg-gray-600 {
  --tw-bg-opacity: 1;
  background-color: rgb(75 85 99 / var(--tw-bg-opacity, 1));
}

.bg-gray-700 {
  --tw-bg-opacity: 1;
  background-color: rgb(55 65 81 / var(--tw-bg-opacity, 1));
}

.bg-gray-800 {
  --tw-bg-opacity: 1;
  background-color: rgb(31 41 55 / var(--tw-bg-opacity, 1));
}

.bg-gray-900 {
  --tw-bg-opacity: 1;
  background-color: rgb(17 24 39 / var(--tw-bg-opacity, 1));
}

.bg-green-100 {
  --tw-bg-opacity: 1;
  background-color: rgb(220 252 231 / var(--tw-bg-opacity, 1));
}

.bg-green-50 {
  --tw-bg-opacity: 1;
  background-color: rgb(240 253 244 / var(--tw-bg-opacity, 1));
}

.bg-green-600 {
  --tw-bg-opacity: 1;
  background-color: rgb(22 163 74 / var(--tw-bg-opacity, 1));
}

.bg-indigo-600\\/40 {
  background-color: rgb(79 70 229 / 0.4);
}

.bg-muted {
  background-color: hsl(var(--muted));
}

.bg-muted\\/30 {
  background-color: hsl(var(--muted) / 0.3);
}

.bg-muted\\/50 {
  background-color: hsl(var(--muted) / 0.5);
}

.bg-orange-100 {
  --tw-bg-opacity: 1;
  background-color: rgb(255 237 213 / var(--tw-bg-opacity, 1));
}

.bg-orange-50 {
  --tw-bg-opacity: 1;
  background-color: rgb(255 247 237 / var(--tw-bg-opacity, 1));
}

.bg-popover {
  background-color: hsl(var(--popover));
}

.bg-primary {
  background-color: hsl(var(--primary));
}

.bg-purple-100 {
  --tw-bg-opacity: 1;
  background-color: rgb(243 232 255 / var(--tw-bg-opacity, 1));
}

.bg-purple-50 {
  --tw-bg-opacity: 1;
  background-color: rgb(250 245 255 / var(--tw-bg-opacity, 1));
}

.bg-red-100 {
  --tw-bg-opacity: 1;
  background-color: rgb(254 226 226 / var(--tw-bg-opacity, 1));
}

.bg-red-50 {
  --tw-bg-opacity: 1;
  background-color: rgb(254 242 242 / var(--tw-bg-opacity, 1));
}

.bg-secondary {
  background-color: hsl(var(--secondary));
}

.bg-sidebar {
  background-color: hsl(var(--sidebar-background));
}

.bg-sidebar-border {
  background-color: hsl(var(--sidebar-border));
}

.bg-teal-50 {
  --tw-bg-opacity: 1;
  background-color: rgb(240 253 250 / var(--tw-bg-opacity, 1));
}

.bg-transparent {
  background-color: transparent;
}

.bg-white {
  --tw-bg-opacity: 1;
  background-color: rgb(255 255 255 / var(--tw-bg-opacity, 1));
}

.bg-white\\/10 {
  background-color: rgb(255 255 255 / 0.1);
}

.bg-gradient-to-br {
  background-image: linear-gradient(to bottom right, var(--tw-gradient-stops));
}

.bg-gradient-to-r {
  background-image: linear-gradient(to right, var(--tw-gradient-stops));
}

.from-blue-50 {
  --tw-gradient-from: #eff6ff var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(239 246 255 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-blue-50\\/50 {
  --tw-gradient-from: rgb(239 246 255 / 0.5) var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(239 246 255 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-green-50 {
  --tw-gradient-from: #f0fdf4 var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(240 253 244 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-indigo-100 {
  --tw-gradient-from: #e0e7ff var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(224 231 255 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-purple-100 {
  --tw-gradient-from: #f3e8ff var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(243 232 255 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-purple-50 {
  --tw-gradient-from: #faf5ff var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(250 245 255 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-purple-50\\/50 {
  --tw-gradient-from: rgb(250 245 255 / 0.5) var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(250 245 255 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-purple-600 {
  --tw-gradient-from: #9333ea var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(147 51 234 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.from-purple-700 {
  --tw-gradient-from: #7e22ce var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(126 34 206 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.via-white {
  --tw-gradient-to: rgb(255 255 255 / 0)  var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), #fff var(--tw-gradient-via-position), var(--tw-gradient-to);
}

.to-blue-100 {
  --tw-gradient-to: #dbeafe var(--tw-gradient-to-position);
}

.to-emerald-50 {
  --tw-gradient-to: #ecfdf5 var(--tw-gradient-to-position);
}

.to-indigo-100 {
  --tw-gradient-to: #e0e7ff var(--tw-gradient-to-position);
}

.to-indigo-50 {
  --tw-gradient-to: #eef2ff var(--tw-gradient-to-position);
}

.to-indigo-50\\/50 {
  --tw-gradient-to: rgb(238 242 255 / 0.5) var(--tw-gradient-to-position);
}

.to-indigo-600 {
  --tw-gradient-to: #4f46e5 var(--tw-gradient-to-position);
}

.to-indigo-700 {
  --tw-gradient-to: #4338ca var(--tw-gradient-to-position);
}

.to-teal-50 {
  --tw-gradient-to: #f0fdfa var(--tw-gradient-to-position);
}

.bg-clip-text {
  -webkit-background-clip: text;
          background-clip: text;
}

.fill-current {
  fill: currentColor;
}

.p-0 {
  padding: 0px;
}

.p-1 {
  padding: 0.25rem;
}

.p-2 {
  padding: 0.5rem;
}

.p-3 {
  padding: 0.75rem;
}

.p-4 {
  padding: 1rem;
}

.p-5 {
  padding: 1.25rem;
}

.p-6 {
  padding: 1.5rem;
}

.p-8 {
  padding: 2rem;
}

.p-\\[1px\\] {
  padding: 1px;
}

.px-1 {
  padding-left: 0.25rem;
  padding-right: 0.25rem;
}

.px-2 {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.px-2\\.5 {
  padding-left: 0.625rem;
  padding-right: 0.625rem;
}

.px-3 {
  padding-left: 0.75rem;
  padding-right: 0.75rem;
}

.px-4 {
  padding-left: 1rem;
  padding-right: 1rem;
}

.px-5 {
  padding-left: 1.25rem;
  padding-right: 1.25rem;
}

.px-6 {
  padding-left: 1.5rem;
  padding-right: 1.5rem;
}

.px-8 {
  padding-left: 2rem;
  padding-right: 2rem;
}

.py-0\\.5 {
  padding-top: 0.125rem;
  padding-bottom: 0.125rem;
}

.py-1 {
  padding-top: 0.25rem;
  padding-bottom: 0.25rem;
}

.py-1\\.5 {
  padding-top: 0.375rem;
  padding-bottom: 0.375rem;
}

.py-12 {
  padding-top: 3rem;
  padding-bottom: 3rem;
}

.py-2 {
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
}

.py-3 {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
}

.py-4 {
  padding-top: 1rem;
  padding-bottom: 1rem;
}

.py-6 {
  padding-top: 1.5rem;
  padding-bottom: 1.5rem;
}

.py-8 {
  padding-top: 2rem;
  padding-bottom: 2rem;
}

.pb-2 {
  padding-bottom: 0.5rem;
}

.pb-3 {
  padding-bottom: 0.75rem;
}

.pb-4 {
  padding-bottom: 1rem;
}

.pb-6 {
  padding-bottom: 1.5rem;
}

.pl-2\\.5 {
  padding-left: 0.625rem;
}

.pl-4 {
  padding-left: 1rem;
}

.pl-8 {
  padding-left: 2rem;
}

.pr-12 {
  padding-right: 3rem;
}

.pr-2 {
  padding-right: 0.5rem;
}

.pr-2\\.5 {
  padding-right: 0.625rem;
}

.pr-8 {
  padding-right: 2rem;
}

.pt-0 {
  padding-top: 0px;
}

.pt-1 {
  padding-top: 0.25rem;
}

.pt-2 {
  padding-top: 0.5rem;
}

.pt-3 {
  padding-top: 0.75rem;
}

.pt-4 {
  padding-top: 1rem;
}

.pt-6 {
  padding-top: 1.5rem;
}

.text-left {
  text-align: left;
}

.text-center {
  text-align: center;
}

.text-right {
  text-align: right;
}

.align-middle {
  vertical-align: middle;
}

.font-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

.text-2xl {
  font-size: 1.5rem;
  line-height: 2rem;
}

.text-3xl {
  font-size: 1.875rem;
  line-height: 2.25rem;
}

.text-\\[0\\.8rem\\] {
  font-size: 0.8rem;
}

.text-base {
  font-size: 1rem;
  line-height: 1.5rem;
}

.text-lg {
  font-size: 1.125rem;
  line-height: 1.75rem;
}

.text-sm {
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.text-xl {
  font-size: 1.25rem;
  line-height: 1.75rem;
}

.text-xs {
  font-size: 0.75rem;
  line-height: 1rem;
}

.font-bold {
  font-weight: 700;
}

.font-medium {
  font-weight: 500;
}

.font-normal {
  font-weight: 400;
}

.font-semibold {
  font-weight: 600;
}

.capitalize {
  text-transform: capitalize;
}

.tabular-nums {
  --tw-numeric-spacing: tabular-nums;
  font-variant-numeric: var(--tw-ordinal) var(--tw-slashed-zero) var(--tw-numeric-figure) var(--tw-numeric-spacing) var(--tw-numeric-fraction);
}

.leading-none {
  line-height: 1;
}

.leading-relaxed {
  line-height: 1.625;
}

.tracking-tight {
  letter-spacing: -0.025em;
}

.tracking-widest {
  letter-spacing: 0.1em;
}

.text-accent-foreground {
  color: hsl(var(--accent-foreground));
}

.text-blue-600 {
  --tw-text-opacity: 1;
  color: rgb(37 99 235 / var(--tw-text-opacity, 1));
}

.text-blue-700 {
  --tw-text-opacity: 1;
  color: rgb(29 78 216 / var(--tw-text-opacity, 1));
}

.text-blue-800 {
  --tw-text-opacity: 1;
  color: rgb(30 64 175 / var(--tw-text-opacity, 1));
}

.text-blue-900 {
  --tw-text-opacity: 1;
  color: rgb(30 58 138 / var(--tw-text-opacity, 1));
}

.text-card-foreground {
  color: hsl(var(--card-foreground));
}

.text-current {
  color: currentColor;
}

.text-destructive {
  color: hsl(var(--destructive));
}

.text-destructive-foreground {
  color: hsl(var(--destructive-foreground));
}

.text-foreground {
  color: hsl(var(--foreground));
}

.text-foreground\\/50 {
  color: hsl(var(--foreground) / 0.5);
}

.text-gray-300 {
  --tw-text-opacity: 1;
  color: rgb(209 213 219 / var(--tw-text-opacity, 1));
}

.text-gray-400 {
  --tw-text-opacity: 1;
  color: rgb(156 163 175 / var(--tw-text-opacity, 1));
}

.text-gray-500 {
  --tw-text-opacity: 1;
  color: rgb(107 114 128 / var(--tw-text-opacity, 1));
}

.text-gray-600 {
  --tw-text-opacity: 1;
  color: rgb(75 85 99 / var(--tw-text-opacity, 1));
}

.text-gray-700 {
  --tw-text-opacity: 1;
  color: rgb(55 65 81 / var(--tw-text-opacity, 1));
}

.text-gray-800 {
  --tw-text-opacity: 1;
  color: rgb(31 41 55 / var(--tw-text-opacity, 1));
}

.text-gray-900 {
  --tw-text-opacity: 1;
  color: rgb(17 24 39 / var(--tw-text-opacity, 1));
}

.text-green-400 {
  --tw-text-opacity: 1;
  color: rgb(74 222 128 / var(--tw-text-opacity, 1));
}

.text-green-600 {
  --tw-text-opacity: 1;
  color: rgb(22 163 74 / var(--tw-text-opacity, 1));
}

.text-green-700 {
  --tw-text-opacity: 1;
  color: rgb(21 128 61 / var(--tw-text-opacity, 1));
}

.text-green-800 {
  --tw-text-opacity: 1;
  color: rgb(22 101 52 / var(--tw-text-opacity, 1));
}

.text-green-900 {
  --tw-text-opacity: 1;
  color: rgb(20 83 45 / var(--tw-text-opacity, 1));
}

.text-indigo-600 {
  --tw-text-opacity: 1;
  color: rgb(79 70 229 / var(--tw-text-opacity, 1));
}

.text-muted-foreground {
  color: hsl(var(--muted-foreground));
}

.text-orange-600 {
  --tw-text-opacity: 1;
  color: rgb(234 88 12 / var(--tw-text-opacity, 1));
}

.text-orange-700 {
  --tw-text-opacity: 1;
  color: rgb(194 65 12 / var(--tw-text-opacity, 1));
}

.text-orange-800 {
  --tw-text-opacity: 1;
  color: rgb(154 52 18 / var(--tw-text-opacity, 1));
}

.text-popover-foreground {
  color: hsl(var(--popover-foreground));
}

.text-primary {
  color: hsl(var(--primary));
}

.text-primary-foreground {
  color: hsl(var(--primary-foreground));
}

.text-purple-100 {
  --tw-text-opacity: 1;
  color: rgb(243 232 255 / var(--tw-text-opacity, 1));
}

.text-purple-600 {
  --tw-text-opacity: 1;
  color: rgb(147 51 234 / var(--tw-text-opacity, 1));
}

.text-purple-700 {
  --tw-text-opacity: 1;
  color: rgb(126 34 206 / var(--tw-text-opacity, 1));
}

.text-red-500 {
  --tw-text-opacity: 1;
  color: rgb(239 68 68 / var(--tw-text-opacity, 1));
}

.text-red-600 {
  --tw-text-opacity: 1;
  color: rgb(220 38 38 / var(--tw-text-opacity, 1));
}

.text-red-700 {
  --tw-text-opacity: 1;
  color: rgb(185 28 28 / var(--tw-text-opacity, 1));
}

.text-red-800 {
  --tw-text-opacity: 1;
  color: rgb(153 27 27 / var(--tw-text-opacity, 1));
}

.text-secondary-foreground {
  color: hsl(var(--secondary-foreground));
}

.text-sidebar-foreground {
  color: hsl(var(--sidebar-foreground));
}

.text-sidebar-foreground\\/70 {
  color: hsl(var(--sidebar-foreground) / 0.7);
}

.text-teal-600 {
  --tw-text-opacity: 1;
  color: rgb(13 148 136 / var(--tw-text-opacity, 1));
}

.text-transparent {
  color: transparent;
}

.text-white {
  --tw-text-opacity: 1;
  color: rgb(255 255 255 / var(--tw-text-opacity, 1));
}

.underline-offset-4 {
  text-underline-offset: 4px;
}

.antialiased {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.opacity-0 {
  opacity: 0;
}

.opacity-50 {
  opacity: 0.5;
}

.opacity-60 {
  opacity: 0.6;
}

.opacity-70 {
  opacity: 0.7;
}

.opacity-90 {
  opacity: 0.9;
}

.shadow {
  --tw-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 1px 3px 0 var(--tw-shadow-color), 0 1px 2px -1px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.shadow-\\[0_0_0_1px_hsl\\(var\\(--sidebar-border\\)\\)\\] {
  --tw-shadow: 0 0 0 1px hsl(var(--sidebar-border));
  --tw-shadow-colored: 0 0 0 1px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.shadow-lg {
  --tw-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 10px 15px -3px var(--tw-shadow-color), 0 4px 6px -4px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.shadow-md {
  --tw-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 4px 6px -1px var(--tw-shadow-color), 0 2px 4px -2px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.shadow-none {
  --tw-shadow: 0 0 #0000;
  --tw-shadow-colored: 0 0 #0000;
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.shadow-sm {
  --tw-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --tw-shadow-colored: 0 1px 2px 0 var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.shadow-xl {
  --tw-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 20px 25px -5px var(--tw-shadow-color), 0 8px 10px -6px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.outline-none {
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.outline {
  outline-style: solid;
}

.ring-0 {
  --tw-ring-offset-shadow: var(--tw-ring-inset) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color);
  --tw-ring-shadow: var(--tw-ring-inset) 0 0 0 calc(0px + var(--tw-ring-offset-width)) var(--tw-ring-color);
  box-shadow: var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow, 0 0 #0000);
}

.ring-2 {
  --tw-ring-offset-shadow: var(--tw-ring-inset) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color);
  --tw-ring-shadow: var(--tw-ring-inset) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color);
  box-shadow: var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow, 0 0 #0000);
}

.ring-ring {
  --tw-ring-color: hsl(var(--ring));
}

.ring-sidebar-ring {
  --tw-ring-color: hsl(var(--sidebar-ring));
}

.ring-offset-background {
  --tw-ring-offset-color: hsl(var(--background));
}

.grayscale {
  --tw-grayscale: grayscale(100%);
  filter: var(--tw-blur) var(--tw-brightness) var(--tw-contrast) var(--tw-grayscale) var(--tw-hue-rotate) var(--tw-invert) var(--tw-saturate) var(--tw-sepia) var(--tw-drop-shadow);
}

.filter {
  filter: var(--tw-blur) var(--tw-brightness) var(--tw-contrast) var(--tw-grayscale) var(--tw-hue-rotate) var(--tw-invert) var(--tw-saturate) var(--tw-sepia) var(--tw-drop-shadow);
}

.transition {
  transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, -webkit-backdrop-filter;
  transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter;
  transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter, -webkit-backdrop-filter;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-\\[left\\2c right\\2c width\\] {
  transition-property: left,right,width;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-\\[margin\\2c opa\\] {
  transition-property: margin,opa;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-\\[width\\2c height\\2c padding\\] {
  transition-property: width,height,padding;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-\\[width\\] {
  transition-property: width;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-all {
  transition-property: all;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-colors {
  transition-property: color, background-color, border-color, text-decoration-color, fill, stroke;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-opacity {
  transition-property: opacity;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-shadow {
  transition-property: box-shadow;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.transition-transform {
  transition-property: transform;
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
  transition-duration: 150ms;
}

.duration-1000 {
  transition-duration: 1000ms;
}

.duration-200 {
  transition-duration: 200ms;
}

.ease-in-out {
  transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}

.ease-linear {
  transition-timing-function: linear;
}

@keyframes enter {
  from {
    opacity: var(--tw-enter-opacity, 1);
    transform: translate3d(var(--tw-enter-translate-x, 0), var(--tw-enter-translate-y, 0), 0) scale3d(var(--tw-enter-scale, 1), var(--tw-enter-scale, 1), var(--tw-enter-scale, 1)) rotate(var(--tw-enter-rotate, 0));
  }
}

@keyframes exit {
  to {
    opacity: var(--tw-exit-opacity, 1);
    transform: translate3d(var(--tw-exit-translate-x, 0), var(--tw-exit-translate-y, 0), 0) scale3d(var(--tw-exit-scale, 1), var(--tw-exit-scale, 1), var(--tw-exit-scale, 1)) rotate(var(--tw-exit-rotate, 0));
  }
}

.animate-in {
  animation-name: enter;
  animation-duration: 150ms;
  --tw-enter-opacity: initial;
  --tw-enter-scale: initial;
  --tw-enter-rotate: initial;
  --tw-enter-translate-x: initial;
  --tw-enter-translate-y: initial;
}

.fade-in-0 {
  --tw-enter-opacity: 0;
}

.fade-in-80 {
  --tw-enter-opacity: 0.8;
}

.zoom-in-95 {
  --tw-enter-scale: .95;
}

.duration-1000 {
  animation-duration: 1000ms;
}

.duration-200 {
  animation-duration: 200ms;
}

.ease-in-out {
  animation-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
}

.ease-linear {
  animation-timing-function: linear;
}

/* Definition of the design system. All colors, gradients, fonts, etc should be defined here. */

/* Global styles for Radix UI portals rendered outside Shadow DOM */

[data-radix-popper-content-wrapper] .hover\\:shadow-md:hover {
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1) !important;
}

[data-radix-popper-content-wrapper] .group[data-variant="floating"] .group-data-\\[variant\\=floating\\]\\:border {
  border-color: hsl(var(--border)) !important;
}

.file\\:border-0::file-selector-button {
  border-width: 0px;
}

.file\\:bg-transparent::file-selector-button {
  background-color: transparent;
}

.file\\:text-sm::file-selector-button {
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.file\\:font-medium::file-selector-button {
  font-weight: 500;
}

.file\\:text-foreground::file-selector-button {
  color: hsl(var(--foreground));
}

.placeholder\\:text-muted-foreground::-moz-placeholder {
  color: hsl(var(--muted-foreground));
}

.placeholder\\:text-muted-foreground::placeholder {
  color: hsl(var(--muted-foreground));
}

.after\\:absolute::after {
  content: var(--tw-content);
  position: absolute;
}

.after\\:-inset-2::after {
  content: var(--tw-content);
  inset: -0.5rem;
}

.after\\:inset-y-0::after {
  content: var(--tw-content);
  top: 0px;
  bottom: 0px;
}

.after\\:left-1\\/2::after {
  content: var(--tw-content);
  left: 50%;
}

.after\\:w-1::after {
  content: var(--tw-content);
  width: 0.25rem;
}

.after\\:w-\\[2px\\]::after {
  content: var(--tw-content);
  width: 2px;
}

.after\\:-translate-x-1\\/2::after {
  content: var(--tw-content);
  --tw-translate-x: -50%;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.first\\:rounded-l-md:first-child {
  border-top-left-radius: calc(var(--radius) - 2px);
  border-bottom-left-radius: calc(var(--radius) - 2px);
}

.first\\:border-l:first-child {
  border-left-width: 1px;
}

.last\\:rounded-r-md:last-child {
  border-top-right-radius: calc(var(--radius) - 2px);
  border-bottom-right-radius: calc(var(--radius) - 2px);
}

.focus-within\\:relative:focus-within {
  position: relative;
}

.focus-within\\:z-20:focus-within {
  z-index: 20;
}

.hover\\:-translate-y-0\\.5:hover {
  --tw-translate-y: -0.125rem;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.hover\\:border-gray-400:hover {
  --tw-border-opacity: 1;
  border-color: rgb(156 163 175 / var(--tw-border-opacity, 1));
}

.hover\\:bg-accent:hover {
  background-color: hsl(var(--accent));
}

.hover\\:bg-blue-700:hover {
  --tw-bg-opacity: 1;
  background-color: rgb(29 78 216 / var(--tw-bg-opacity, 1));
}

.hover\\:bg-destructive\\/80:hover {
  background-color: hsl(var(--destructive) / 0.8);
}

.hover\\:bg-destructive\\/90:hover {
  background-color: hsl(var(--destructive) / 0.9);
}

.hover\\:bg-gray-50:hover {
  --tw-bg-opacity: 1;
  background-color: rgb(249 250 251 / var(--tw-bg-opacity, 1));
}

.hover\\:bg-gray-700:hover {
  --tw-bg-opacity: 1;
  background-color: rgb(55 65 81 / var(--tw-bg-opacity, 1));
}

.hover\\:bg-gray-800:hover {
  --tw-bg-opacity: 1;
  background-color: rgb(31 41 55 / var(--tw-bg-opacity, 1));
}

.hover\\:bg-green-50:hover {
  --tw-bg-opacity: 1;
  background-color: rgb(240 253 244 / var(--tw-bg-opacity, 1));
}

.hover\\:bg-muted:hover {
  background-color: hsl(var(--muted));
}

.hover\\:bg-muted\\/50:hover {
  background-color: hsl(var(--muted) / 0.5);
}

.hover\\:bg-primary:hover {
  background-color: hsl(var(--primary));
}

.hover\\:bg-primary\\/80:hover {
  background-color: hsl(var(--primary) / 0.8);
}

.hover\\:bg-primary\\/90:hover {
  background-color: hsl(var(--primary) / 0.9);
}

.hover\\:bg-red-100:hover {
  --tw-bg-opacity: 1;
  background-color: rgb(254 226 226 / var(--tw-bg-opacity, 1));
}

.hover\\:bg-secondary:hover {
  background-color: hsl(var(--secondary));
}

.hover\\:bg-secondary\\/80:hover {
  background-color: hsl(var(--secondary) / 0.8);
}

.hover\\:bg-sidebar-accent:hover {
  background-color: hsl(var(--sidebar-accent));
}

.hover\\:text-accent-foreground:hover {
  color: hsl(var(--accent-foreground));
}

.hover\\:text-blue-800:hover {
  --tw-text-opacity: 1;
  color: rgb(30 64 175 / var(--tw-text-opacity, 1));
}

.hover\\:text-foreground:hover {
  color: hsl(var(--foreground));
}

.hover\\:text-muted-foreground:hover {
  color: hsl(var(--muted-foreground));
}

.hover\\:text-primary-foreground:hover {
  color: hsl(var(--primary-foreground));
}

.hover\\:text-sidebar-accent-foreground:hover {
  color: hsl(var(--sidebar-accent-foreground));
}

.hover\\:text-white:hover {
  --tw-text-opacity: 1;
  color: rgb(255 255 255 / var(--tw-text-opacity, 1));
}

.hover\\:underline:hover {
  text-decoration-line: underline;
}

.hover\\:opacity-100:hover {
  opacity: 1;
}

.hover\\:shadow-\\[0_0_0_1px_hsl\\(var\\(--sidebar-accent\\)\\)\\]:hover {
  --tw-shadow: 0 0 0 1px hsl(var(--sidebar-accent));
  --tw-shadow-colored: 0 0 0 1px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.hover\\:shadow-md:hover {
  --tw-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 4px 6px -1px var(--tw-shadow-color), 0 2px 4px -2px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.hover\\:shadow-sm:hover {
  --tw-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --tw-shadow-colored: 0 1px 2px 0 var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.hover\\:shadow-xl:hover {
  --tw-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 20px 25px -5px var(--tw-shadow-color), 0 8px 10px -6px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.hover\\:after\\:bg-sidebar-border:hover::after {
  content: var(--tw-content);
  background-color: hsl(var(--sidebar-border));
}

.focus\\:bg-accent:focus {
  background-color: hsl(var(--accent));
}

.focus\\:bg-primary:focus {
  background-color: hsl(var(--primary));
}

.focus\\:text-accent-foreground:focus {
  color: hsl(var(--accent-foreground));
}

.focus\\:text-primary-foreground:focus {
  color: hsl(var(--primary-foreground));
}

.focus\\:opacity-100:focus {
  opacity: 1;
}

.focus\\:outline-none:focus {
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.focus\\:ring-2:focus {
  --tw-ring-offset-shadow: var(--tw-ring-inset) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color);
  --tw-ring-shadow: var(--tw-ring-inset) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color);
  box-shadow: var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow, 0 0 #0000);
}

.focus\\:ring-ring:focus {
  --tw-ring-color: hsl(var(--ring));
}

.focus\\:ring-offset-2:focus {
  --tw-ring-offset-width: 2px;
}

.focus-visible\\:outline-none:focus-visible {
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.focus-visible\\:ring-1:focus-visible {
  --tw-ring-offset-shadow: var(--tw-ring-inset) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color);
  --tw-ring-shadow: var(--tw-ring-inset) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color);
  box-shadow: var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow, 0 0 #0000);
}

.focus-visible\\:ring-2:focus-visible {
  --tw-ring-offset-shadow: var(--tw-ring-inset) 0 0 0 var(--tw-ring-offset-width) var(--tw-ring-offset-color);
  --tw-ring-shadow: var(--tw-ring-inset) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color);
  box-shadow: var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow, 0 0 #0000);
}

.focus-visible\\:ring-ring:focus-visible {
  --tw-ring-color: hsl(var(--ring));
}

.focus-visible\\:ring-sidebar-ring:focus-visible {
  --tw-ring-color: hsl(var(--sidebar-ring));
}

.focus-visible\\:ring-offset-1:focus-visible {
  --tw-ring-offset-width: 1px;
}

.focus-visible\\:ring-offset-2:focus-visible {
  --tw-ring-offset-width: 2px;
}

.focus-visible\\:ring-offset-background:focus-visible {
  --tw-ring-offset-color: hsl(var(--background));
}

.active\\:bg-sidebar-accent:active {
  background-color: hsl(var(--sidebar-accent));
}

.active\\:text-sidebar-accent-foreground:active {
  color: hsl(var(--sidebar-accent-foreground));
}

.disabled\\:pointer-events-none:disabled {
  pointer-events: none;
}

.disabled\\:cursor-not-allowed:disabled {
  cursor: not-allowed;
}

.disabled\\:opacity-50:disabled {
  opacity: 0.5;
}

.group\\/menu-item:focus-within .group-focus-within\\/menu-item\\:opacity-100 {
  opacity: 1;
}

.group\\/menu-item:hover .group-hover\\/menu-item\\:opacity-100 {
  opacity: 1;
}

.group:hover .group-hover\\:opacity-100 {
  opacity: 1;
}

.group.destructive .group-\\[\\.destructive\\]\\:border-muted\\/40 {
  border-color: hsl(var(--muted) / 0.4);
}

.group.toaster .group-\\[\\.toaster\\]\\:border-border {
  border-color: hsl(var(--border));
}

.group.toast .group-\\[\\.toast\\]\\:bg-muted {
  background-color: hsl(var(--muted));
}

.group.toast .group-\\[\\.toast\\]\\:bg-primary {
  background-color: hsl(var(--primary));
}

.group.toaster .group-\\[\\.toaster\\]\\:bg-background {
  background-color: hsl(var(--background));
}

.group.destructive .group-\\[\\.destructive\\]\\:text-red-300 {
  --tw-text-opacity: 1;
  color: rgb(252 165 165 / var(--tw-text-opacity, 1));
}

.group.toast .group-\\[\\.toast\\]\\:text-muted-foreground {
  color: hsl(var(--muted-foreground));
}

.group.toast .group-\\[\\.toast\\]\\:text-primary-foreground {
  color: hsl(var(--primary-foreground));
}

.group.toaster .group-\\[\\.toaster\\]\\:text-foreground {
  color: hsl(var(--foreground));
}

.group.toaster .group-\\[\\.toaster\\]\\:shadow-lg {
  --tw-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 10px 15px -3px var(--tw-shadow-color), 0 4px 6px -4px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.group.destructive .group-\\[\\.destructive\\]\\:hover\\:border-destructive\\/30:hover {
  border-color: hsl(var(--destructive) / 0.3);
}

.group.destructive .group-\\[\\.destructive\\]\\:hover\\:bg-destructive:hover {
  background-color: hsl(var(--destructive));
}

.group.destructive .group-\\[\\.destructive\\]\\:hover\\:text-destructive-foreground:hover {
  color: hsl(var(--destructive-foreground));
}

.group.destructive .group-\\[\\.destructive\\]\\:hover\\:text-red-50:hover {
  --tw-text-opacity: 1;
  color: rgb(254 242 242 / var(--tw-text-opacity, 1));
}

.group.destructive .group-\\[\\.destructive\\]\\:focus\\:ring-destructive:focus {
  --tw-ring-color: hsl(var(--destructive));
}

.group.destructive .group-\\[\\.destructive\\]\\:focus\\:ring-red-400:focus {
  --tw-ring-opacity: 1;
  --tw-ring-color: rgb(248 113 113 / var(--tw-ring-opacity, 1));
}

.group.destructive .group-\\[\\.destructive\\]\\:focus\\:ring-offset-red-600:focus {
  --tw-ring-offset-color: #dc2626;
}

.peer\\/menu-button:hover ~ .peer-hover\\/menu-button\\:text-sidebar-accent-foreground {
  color: hsl(var(--sidebar-accent-foreground));
}

.peer:disabled ~ .peer-disabled\\:cursor-not-allowed {
  cursor: not-allowed;
}

.peer:disabled ~ .peer-disabled\\:opacity-70 {
  opacity: 0.7;
}

.has-\\[\\[data-variant\\=inset\\]\\]\\:bg-sidebar:has([data-variant=inset]) {
  background-color: hsl(var(--sidebar-background));
}

.has-\\[\\:disabled\\]\\:opacity-50:has(:disabled) {
  opacity: 0.5;
}

.group\\/menu-item:has([data-sidebar=menu-action]) .group-has-\\[\\[data-sidebar\\=menu-action\\]\\]\\/menu-item\\:pr-8 {
  padding-right: 2rem;
}

.aria-disabled\\:pointer-events-none[aria-disabled="true"] {
  pointer-events: none;
}

.aria-disabled\\:opacity-50[aria-disabled="true"] {
  opacity: 0.5;
}

.aria-selected\\:bg-accent[aria-selected="true"] {
  background-color: hsl(var(--accent));
}

.aria-selected\\:bg-accent\\/50[aria-selected="true"] {
  background-color: hsl(var(--accent) / 0.5);
}

.aria-selected\\:text-accent-foreground[aria-selected="true"] {
  color: hsl(var(--accent-foreground));
}

.aria-selected\\:text-muted-foreground[aria-selected="true"] {
  color: hsl(var(--muted-foreground));
}

.aria-selected\\:opacity-100[aria-selected="true"] {
  opacity: 1;
}

.aria-selected\\:opacity-30[aria-selected="true"] {
  opacity: 0.3;
}

.data-\\[disabled\\=true\\]\\:pointer-events-none[data-disabled="true"] {
  pointer-events: none;
}

.data-\\[disabled\\]\\:pointer-events-none[data-disabled] {
  pointer-events: none;
}

.data-\\[panel-group-direction\\=vertical\\]\\:h-px[data-panel-group-direction="vertical"] {
  height: 1px;
}

.data-\\[panel-group-direction\\=vertical\\]\\:w-full[data-panel-group-direction="vertical"] {
  width: 100%;
}

.data-\\[side\\=bottom\\]\\:translate-y-1[data-side="bottom"] {
  --tw-translate-y: 0.25rem;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[side\\=left\\]\\:-translate-x-1[data-side="left"] {
  --tw-translate-x: -0.25rem;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[side\\=right\\]\\:translate-x-1[data-side="right"] {
  --tw-translate-x: 0.25rem;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[side\\=top\\]\\:-translate-y-1[data-side="top"] {
  --tw-translate-y: -0.25rem;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[state\\=checked\\]\\:translate-x-5[data-state="checked"] {
  --tw-translate-x: 1.25rem;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[state\\=unchecked\\]\\:translate-x-0[data-state="unchecked"] {
  --tw-translate-x: 0px;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[swipe\\=cancel\\]\\:translate-x-0[data-swipe="cancel"] {
  --tw-translate-x: 0px;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[swipe\\=end\\]\\:translate-x-\\[var\\(--radix-toast-swipe-end-x\\)\\][data-swipe="end"] {
  --tw-translate-x: var(--radix-toast-swipe-end-x);
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[swipe\\=move\\]\\:translate-x-\\[var\\(--radix-toast-swipe-move-x\\)\\][data-swipe="move"] {
  --tw-translate-x: var(--radix-toast-swipe-move-x);
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

@keyframes accordion-up {
  from {
    height: var(--radix-accordion-content-height);
  }

  to {
    height: 0;
  }
}

.data-\\[state\\=closed\\]\\:animate-accordion-up[data-state="closed"] {
  animation: accordion-up 0.2s ease-out;
}

@keyframes accordion-down {
  from {
    height: 0;
  }

  to {
    height: var(--radix-accordion-content-height);
  }
}

.data-\\[state\\=open\\]\\:animate-accordion-down[data-state="open"] {
  animation: accordion-down 0.2s ease-out;
}

.data-\\[panel-group-direction\\=vertical\\]\\:flex-col[data-panel-group-direction="vertical"] {
  flex-direction: column;
}

.data-\\[state\\=active\\]\\:border-b-2[data-state="active"] {
  border-bottom-width: 2px;
}

.data-\\[state\\=active\\]\\:border-purple-600[data-state="active"] {
  --tw-border-opacity: 1;
  border-color: rgb(147 51 234 / var(--tw-border-opacity, 1));
}

.data-\\[active\\=true\\]\\:bg-sidebar-accent[data-active="true"] {
  background-color: hsl(var(--sidebar-accent));
}

.data-\\[active\\]\\:bg-accent\\/50[data-active] {
  background-color: hsl(var(--accent) / 0.5);
}

.data-\\[selected\\=\\'true\\'\\]\\:bg-accent[data-selected='true'] {
  background-color: hsl(var(--accent));
}

.data-\\[state\\=active\\]\\:bg-background[data-state="active"] {
  background-color: hsl(var(--background));
}

.data-\\[state\\=checked\\]\\:bg-primary[data-state="checked"] {
  background-color: hsl(var(--primary));
}

.data-\\[state\\=on\\]\\:bg-accent[data-state="on"] {
  background-color: hsl(var(--accent));
}

.data-\\[state\\=open\\]\\:bg-accent[data-state="open"] {
  background-color: hsl(var(--accent));
}

.data-\\[state\\=open\\]\\:bg-accent\\/50[data-state="open"] {
  background-color: hsl(var(--accent) / 0.5);
}

.data-\\[state\\=open\\]\\:bg-secondary[data-state="open"] {
  background-color: hsl(var(--secondary));
}

.data-\\[state\\=selected\\]\\:bg-muted[data-state="selected"] {
  background-color: hsl(var(--muted));
}

.data-\\[state\\=unchecked\\]\\:bg-input[data-state="unchecked"] {
  background-color: hsl(var(--input));
}

.data-\\[state\\=active\\]\\:bg-gradient-to-r[data-state="active"] {
  background-image: linear-gradient(to right, var(--tw-gradient-stops));
}

.data-\\[state\\=active\\]\\:from-purple-600[data-state="active"] {
  --tw-gradient-from: #9333ea var(--tw-gradient-from-position);
  --tw-gradient-to: rgb(147 51 234 / 0) var(--tw-gradient-to-position);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to);
}

.data-\\[state\\=active\\]\\:to-indigo-600[data-state="active"] {
  --tw-gradient-to: #4f46e5 var(--tw-gradient-to-position);
}

.data-\\[active\\=true\\]\\:font-medium[data-active="true"] {
  font-weight: 500;
}

.data-\\[active\\=true\\]\\:text-sidebar-accent-foreground[data-active="true"] {
  color: hsl(var(--sidebar-accent-foreground));
}

.data-\\[selected\\=true\\]\\:text-accent-foreground[data-selected="true"] {
  color: hsl(var(--accent-foreground));
}

.data-\\[state\\=active\\]\\:text-foreground[data-state="active"] {
  color: hsl(var(--foreground));
}

.data-\\[state\\=active\\]\\:text-white[data-state="active"] {
  --tw-text-opacity: 1;
  color: rgb(255 255 255 / var(--tw-text-opacity, 1));
}

.data-\\[state\\=checked\\]\\:text-primary-foreground[data-state="checked"] {
  color: hsl(var(--primary-foreground));
}

.data-\\[state\\=on\\]\\:text-accent-foreground[data-state="on"] {
  color: hsl(var(--accent-foreground));
}

.data-\\[state\\=open\\]\\:text-accent-foreground[data-state="open"] {
  color: hsl(var(--accent-foreground));
}

.data-\\[state\\=open\\]\\:text-muted-foreground[data-state="open"] {
  color: hsl(var(--muted-foreground));
}

.data-\\[disabled\\=true\\]\\:opacity-50[data-disabled="true"] {
  opacity: 0.5;
}

.data-\\[disabled\\]\\:opacity-50[data-disabled] {
  opacity: 0.5;
}

.data-\\[state\\=open\\]\\:opacity-100[data-state="open"] {
  opacity: 1;
}

.data-\\[state\\=active\\]\\:shadow-lg[data-state="active"] {
  --tw-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 10px 15px -3px var(--tw-shadow-color), 0 4px 6px -4px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.data-\\[state\\=active\\]\\:shadow-sm[data-state="active"] {
  --tw-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --tw-shadow-colored: 0 1px 2px 0 var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.data-\\[swipe\\=move\\]\\:transition-none[data-swipe="move"] {
  transition-property: none;
}

.data-\\[state\\=closed\\]\\:duration-300[data-state="closed"] {
  transition-duration: 300ms;
}

.data-\\[state\\=open\\]\\:duration-500[data-state="open"] {
  transition-duration: 500ms;
}

.data-\\[motion\\^\\=from-\\]\\:animate-in[data-motion^="from-"] {
  animation-name: enter;
  animation-duration: 150ms;
  --tw-enter-opacity: initial;
  --tw-enter-scale: initial;
  --tw-enter-rotate: initial;
  --tw-enter-translate-x: initial;
  --tw-enter-translate-y: initial;
}

.data-\\[state\\=open\\]\\:animate-in[data-state="open"] {
  animation-name: enter;
  animation-duration: 150ms;
  --tw-enter-opacity: initial;
  --tw-enter-scale: initial;
  --tw-enter-rotate: initial;
  --tw-enter-translate-x: initial;
  --tw-enter-translate-y: initial;
}

.data-\\[state\\=visible\\]\\:animate-in[data-state="visible"] {
  animation-name: enter;
  animation-duration: 150ms;
  --tw-enter-opacity: initial;
  --tw-enter-scale: initial;
  --tw-enter-rotate: initial;
  --tw-enter-translate-x: initial;
  --tw-enter-translate-y: initial;
}

.data-\\[motion\\^\\=to-\\]\\:animate-out[data-motion^="to-"] {
  animation-name: exit;
  animation-duration: 150ms;
  --tw-exit-opacity: initial;
  --tw-exit-scale: initial;
  --tw-exit-rotate: initial;
  --tw-exit-translate-x: initial;
  --tw-exit-translate-y: initial;
}

.data-\\[state\\=closed\\]\\:animate-out[data-state="closed"] {
  animation-name: exit;
  animation-duration: 150ms;
  --tw-exit-opacity: initial;
  --tw-exit-scale: initial;
  --tw-exit-rotate: initial;
  --tw-exit-translate-x: initial;
  --tw-exit-translate-y: initial;
}

.data-\\[state\\=hidden\\]\\:animate-out[data-state="hidden"] {
  animation-name: exit;
  animation-duration: 150ms;
  --tw-exit-opacity: initial;
  --tw-exit-scale: initial;
  --tw-exit-rotate: initial;
  --tw-exit-translate-x: initial;
  --tw-exit-translate-y: initial;
}

.data-\\[swipe\\=end\\]\\:animate-out[data-swipe="end"] {
  animation-name: exit;
  animation-duration: 150ms;
  --tw-exit-opacity: initial;
  --tw-exit-scale: initial;
  --tw-exit-rotate: initial;
  --tw-exit-translate-x: initial;
  --tw-exit-translate-y: initial;
}

.data-\\[motion\\^\\=from-\\]\\:fade-in[data-motion^="from-"] {
  --tw-enter-opacity: 0;
}

.data-\\[motion\\^\\=to-\\]\\:fade-out[data-motion^="to-"] {
  --tw-exit-opacity: 0;
}

.data-\\[state\\=closed\\]\\:fade-out-0[data-state="closed"] {
  --tw-exit-opacity: 0;
}

.data-\\[state\\=closed\\]\\:fade-out-80[data-state="closed"] {
  --tw-exit-opacity: 0.8;
}

.data-\\[state\\=hidden\\]\\:fade-out[data-state="hidden"] {
  --tw-exit-opacity: 0;
}

.data-\\[state\\=open\\]\\:fade-in-0[data-state="open"] {
  --tw-enter-opacity: 0;
}

.data-\\[state\\=visible\\]\\:fade-in[data-state="visible"] {
  --tw-enter-opacity: 0;
}

.data-\\[state\\=closed\\]\\:zoom-out-95[data-state="closed"] {
  --tw-exit-scale: .95;
}

.data-\\[state\\=open\\]\\:zoom-in-90[data-state="open"] {
  --tw-enter-scale: .9;
}

.data-\\[state\\=open\\]\\:zoom-in-95[data-state="open"] {
  --tw-enter-scale: .95;
}

.data-\\[motion\\=from-end\\]\\:slide-in-from-right-52[data-motion="from-end"] {
  --tw-enter-translate-x: 13rem;
}

.data-\\[motion\\=from-start\\]\\:slide-in-from-left-52[data-motion="from-start"] {
  --tw-enter-translate-x: -13rem;
}

.data-\\[motion\\=to-end\\]\\:slide-out-to-right-52[data-motion="to-end"] {
  --tw-exit-translate-x: 13rem;
}

.data-\\[motion\\=to-start\\]\\:slide-out-to-left-52[data-motion="to-start"] {
  --tw-exit-translate-x: -13rem;
}

.data-\\[side\\=bottom\\]\\:slide-in-from-top-2[data-side="bottom"] {
  --tw-enter-translate-y: -0.5rem;
}

.data-\\[side\\=left\\]\\:slide-in-from-right-2[data-side="left"] {
  --tw-enter-translate-x: 0.5rem;
}

.data-\\[side\\=right\\]\\:slide-in-from-left-2[data-side="right"] {
  --tw-enter-translate-x: -0.5rem;
}

.data-\\[side\\=top\\]\\:slide-in-from-bottom-2[data-side="top"] {
  --tw-enter-translate-y: 0.5rem;
}

.data-\\[state\\=closed\\]\\:slide-out-to-bottom[data-state="closed"] {
  --tw-exit-translate-y: 100%;
}

.data-\\[state\\=closed\\]\\:slide-out-to-left[data-state="closed"] {
  --tw-exit-translate-x: -100%;
}

.data-\\[state\\=closed\\]\\:slide-out-to-left-1\\/2[data-state="closed"] {
  --tw-exit-translate-x: -50%;
}

.data-\\[state\\=closed\\]\\:slide-out-to-right[data-state="closed"] {
  --tw-exit-translate-x: 100%;
}

.data-\\[state\\=closed\\]\\:slide-out-to-right-full[data-state="closed"] {
  --tw-exit-translate-x: 100%;
}

.data-\\[state\\=closed\\]\\:slide-out-to-top[data-state="closed"] {
  --tw-exit-translate-y: -100%;
}

.data-\\[state\\=closed\\]\\:slide-out-to-top-\\[48\\%\\][data-state="closed"] {
  --tw-exit-translate-y: -48%;
}

.data-\\[state\\=open\\]\\:slide-in-from-bottom[data-state="open"] {
  --tw-enter-translate-y: 100%;
}

.data-\\[state\\=open\\]\\:slide-in-from-left[data-state="open"] {
  --tw-enter-translate-x: -100%;
}

.data-\\[state\\=open\\]\\:slide-in-from-left-1\\/2[data-state="open"] {
  --tw-enter-translate-x: -50%;
}

.data-\\[state\\=open\\]\\:slide-in-from-right[data-state="open"] {
  --tw-enter-translate-x: 100%;
}

.data-\\[state\\=open\\]\\:slide-in-from-top[data-state="open"] {
  --tw-enter-translate-y: -100%;
}

.data-\\[state\\=open\\]\\:slide-in-from-top-\\[48\\%\\][data-state="open"] {
  --tw-enter-translate-y: -48%;
}

.data-\\[state\\=open\\]\\:slide-in-from-top-full[data-state="open"] {
  --tw-enter-translate-y: -100%;
}

.data-\\[state\\=closed\\]\\:duration-300[data-state="closed"] {
  animation-duration: 300ms;
}

.data-\\[state\\=open\\]\\:duration-500[data-state="open"] {
  animation-duration: 500ms;
}

.data-\\[panel-group-direction\\=vertical\\]\\:after\\:left-0[data-panel-group-direction="vertical"]::after {
  content: var(--tw-content);
  left: 0px;
}

.data-\\[panel-group-direction\\=vertical\\]\\:after\\:h-1[data-panel-group-direction="vertical"]::after {
  content: var(--tw-content);
  height: 0.25rem;
}

.data-\\[panel-group-direction\\=vertical\\]\\:after\\:w-full[data-panel-group-direction="vertical"]::after {
  content: var(--tw-content);
  width: 100%;
}

.data-\\[panel-group-direction\\=vertical\\]\\:after\\:-translate-y-1\\/2[data-panel-group-direction="vertical"]::after {
  content: var(--tw-content);
  --tw-translate-y: -50%;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[panel-group-direction\\=vertical\\]\\:after\\:translate-x-0[data-panel-group-direction="vertical"]::after {
  content: var(--tw-content);
  --tw-translate-x: 0px;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.data-\\[state\\=open\\]\\:hover\\:bg-sidebar-accent:hover[data-state="open"] {
  background-color: hsl(var(--sidebar-accent));
}

.data-\\[state\\=open\\]\\:hover\\:text-sidebar-accent-foreground:hover[data-state="open"] {
  color: hsl(var(--sidebar-accent-foreground));
}

.group[data-collapsible="offcanvas"] .group-data-\\[collapsible\\=offcanvas\\]\\:left-\\[calc\\(var\\(--sidebar-width\\)\\*-1\\)\\] {
  left: calc(var(--sidebar-width) * -1);
}

.group[data-collapsible="offcanvas"] .group-data-\\[collapsible\\=offcanvas\\]\\:right-\\[calc\\(var\\(--sidebar-width\\)\\*-1\\)\\] {
  right: calc(var(--sidebar-width) * -1);
}

.group[data-side="left"] .group-data-\\[side\\=left\\]\\:-right-4 {
  right: -1rem;
}

.group[data-side="right"] .group-data-\\[side\\=right\\]\\:left-0 {
  left: 0px;
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:-mt-8 {
  margin-top: -2rem;
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:hidden {
  display: none;
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:\\!size-8 {
  width: 2rem !important;
  height: 2rem !important;
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:w-\\[--sidebar-width-icon\\] {
  width: var(--sidebar-width-icon);
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:w-\\[calc\\(var\\(--sidebar-width-icon\\)_\\+_theme\\(spacing\\.4\\)\\)\\] {
  width: calc(var(--sidebar-width-icon) + 1rem);
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:w-\\[calc\\(var\\(--sidebar-width-icon\\)_\\+_theme\\(spacing\\.4\\)_\\+2px\\)\\] {
  width: calc(var(--sidebar-width-icon) + 1rem + 2px);
}

.group[data-collapsible="offcanvas"] .group-data-\\[collapsible\\=offcanvas\\]\\:w-0 {
  width: 0px;
}

.group[data-collapsible="offcanvas"] .group-data-\\[collapsible\\=offcanvas\\]\\:translate-x-0 {
  --tw-translate-x: 0px;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.group[data-side="right"] .group-data-\\[side\\=right\\]\\:rotate-180 {
  --tw-rotate: 180deg;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.group[data-state="open"] .group-data-\\[state\\=open\\]\\:rotate-180 {
  --tw-rotate: 180deg;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:overflow-hidden {
  overflow: hidden;
}

.group[data-variant="floating"] .group-data-\\[variant\\=floating\\]\\:rounded-lg {
  border-radius: var(--radius);
}

.group[data-variant="floating"] .group-data-\\[variant\\=floating\\]\\:border {
  border-width: 1px;
}

.group[data-side="left"] .group-data-\\[side\\=left\\]\\:border-r {
  border-right-width: 1px;
}

.group[data-side="right"] .group-data-\\[side\\=right\\]\\:border-l {
  border-left-width: 1px;
}

.group[data-variant="floating"] .group-data-\\[variant\\=floating\\]\\:border-sidebar-border {
  border-color: hsl(var(--sidebar-border));
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:\\!p-0 {
  padding: 0px !important;
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:\\!p-2 {
  padding: 0.5rem !important;
}

.group[data-collapsible="icon"] .group-data-\\[collapsible\\=icon\\]\\:opacity-0 {
  opacity: 0;
}

.group[data-variant="floating"] .group-data-\\[variant\\=floating\\]\\:shadow {
  --tw-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
  --tw-shadow-colored: 0 1px 3px 0 var(--tw-shadow-color), 0 1px 2px -1px var(--tw-shadow-color);
  box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
}

.group[data-collapsible="offcanvas"] .group-data-\\[collapsible\\=offcanvas\\]\\:after\\:left-full::after {
  content: var(--tw-content);
  left: 100%;
}

.group[data-collapsible="offcanvas"] .group-data-\\[collapsible\\=offcanvas\\]\\:hover\\:bg-sidebar:hover {
  background-color: hsl(var(--sidebar-background));
}

.peer\\/menu-button[data-size="default"] ~ .peer-data-\\[size\\=default\\]\\/menu-button\\:top-1\\.5 {
  top: 0.375rem;
}

.peer\\/menu-button[data-size="lg"] ~ .peer-data-\\[size\\=lg\\]\\/menu-button\\:top-2\\.5 {
  top: 0.625rem;
}

.peer\\/menu-button[data-size="sm"] ~ .peer-data-\\[size\\=sm\\]\\/menu-button\\:top-1 {
  top: 0.25rem;
}

.peer[data-variant="inset"] ~ .peer-data-\\[variant\\=inset\\]\\:min-h-\\[calc\\(100svh-theme\\(spacing\\.4\\)\\)\\] {
  min-height: calc(100svh - 1rem);
}

.peer\\/menu-button[data-active="true"] ~ .peer-data-\\[active\\=true\\]\\/menu-button\\:text-sidebar-accent-foreground {
  color: hsl(var(--sidebar-accent-foreground));
}

.dark\\:border-destructive:is(.dark *) {
  border-color: hsl(var(--destructive));
}

.dark\\:border-gray-700:is(.dark *) {
  --tw-border-opacity: 1;
  border-color: rgb(55 65 81 / var(--tw-border-opacity, 1));
}

.dark\\:bg-gray-800:is(.dark *) {
  --tw-bg-opacity: 1;
  background-color: rgb(31 41 55 / var(--tw-bg-opacity, 1));
}

.dark\\:text-gray-300:is(.dark *) {
  --tw-text-opacity: 1;
  color: rgb(209 213 219 / var(--tw-text-opacity, 1));
}

.dark\\:hover\\:bg-gray-700:hover:is(.dark *) {
  --tw-bg-opacity: 1;
  background-color: rgb(55 65 81 / var(--tw-bg-opacity, 1));
}

@media (min-width: 640px) {
  .sm\\:bottom-0 {
    bottom: 0px;
  }

  .sm\\:right-0 {
    right: 0px;
  }

  .sm\\:top-auto {
    top: auto;
  }

  .sm\\:col-span-2 {
    grid-column: span 2 / span 2;
  }

  .sm\\:mb-0 {
    margin-bottom: 0px;
  }

  .sm\\:mb-2 {
    margin-bottom: 0.5rem;
  }

  .sm\\:mb-3 {
    margin-bottom: 0.75rem;
  }

  .sm\\:mb-4 {
    margin-bottom: 1rem;
  }

  .sm\\:mb-6 {
    margin-bottom: 1.5rem;
  }

  .sm\\:mb-8 {
    margin-bottom: 2rem;
  }

  .sm\\:mr-4 {
    margin-right: 1rem;
  }

  .sm\\:mt-0 {
    margin-top: 0px;
  }

  .sm\\:mt-4 {
    margin-top: 1rem;
  }

  .sm\\:mt-8 {
    margin-top: 2rem;
  }

  .sm\\:inline {
    display: inline;
  }

  .sm\\:flex {
    display: flex;
  }

  .sm\\:h-4 {
    height: 1rem;
  }

  .sm\\:h-5 {
    height: 1.25rem;
  }

  .sm\\:h-6 {
    height: 1.5rem;
  }

  .sm\\:h-8 {
    height: 2rem;
  }

  .sm\\:w-4 {
    width: 1rem;
  }

  .sm\\:w-5 {
    width: 1.25rem;
  }

  .sm\\:w-6 {
    width: 1.5rem;
  }

  .sm\\:w-8 {
    width: 2rem;
  }

  .sm\\:max-w-sm {
    max-width: 24rem;
  }

  .sm\\:grid-cols-2 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .sm\\:flex-row {
    flex-direction: row;
  }

  .sm\\:flex-col {
    flex-direction: column;
  }

  .sm\\:items-center {
    align-items: center;
  }

  .sm\\:justify-start {
    justify-content: flex-start;
  }

  .sm\\:justify-end {
    justify-content: flex-end;
  }

  .sm\\:justify-between {
    justify-content: space-between;
  }

  .sm\\:gap-2 {
    gap: 0.5rem;
  }

  .sm\\:gap-2\\.5 {
    gap: 0.625rem;
  }

  .sm\\:gap-4 {
    gap: 1rem;
  }

  .sm\\:gap-6 {
    gap: 1.5rem;
  }

  .sm\\:gap-8 {
    gap: 2rem;
  }

  .sm\\:space-x-2 > :not([hidden]) ~ :not([hidden]) {
    --tw-space-x-reverse: 0;
    margin-right: calc(0.5rem * var(--tw-space-x-reverse));
    margin-left: calc(0.5rem * calc(1 - var(--tw-space-x-reverse)));
  }

  .sm\\:space-x-3 > :not([hidden]) ~ :not([hidden]) {
    --tw-space-x-reverse: 0;
    margin-right: calc(0.75rem * var(--tw-space-x-reverse));
    margin-left: calc(0.75rem * calc(1 - var(--tw-space-x-reverse)));
  }

  .sm\\:space-x-4 > :not([hidden]) ~ :not([hidden]) {
    --tw-space-x-reverse: 0;
    margin-right: calc(1rem * var(--tw-space-x-reverse));
    margin-left: calc(1rem * calc(1 - var(--tw-space-x-reverse)));
  }

  .sm\\:space-y-0 > :not([hidden]) ~ :not([hidden]) {
    --tw-space-y-reverse: 0;
    margin-top: calc(0px * calc(1 - var(--tw-space-y-reverse)));
    margin-bottom: calc(0px * var(--tw-space-y-reverse));
  }

  .sm\\:space-y-2 > :not([hidden]) ~ :not([hidden]) {
    --tw-space-y-reverse: 0;
    margin-top: calc(0.5rem * calc(1 - var(--tw-space-y-reverse)));
    margin-bottom: calc(0.5rem * var(--tw-space-y-reverse));
  }

  .sm\\:space-y-3 > :not([hidden]) ~ :not([hidden]) {
    --tw-space-y-reverse: 0;
    margin-top: calc(0.75rem * calc(1 - var(--tw-space-y-reverse)));
    margin-bottom: calc(0.75rem * var(--tw-space-y-reverse));
  }

  .sm\\:space-y-8 > :not([hidden]) ~ :not([hidden]) {
    --tw-space-y-reverse: 0;
    margin-top: calc(2rem * calc(1 - var(--tw-space-y-reverse)));
    margin-bottom: calc(2rem * var(--tw-space-y-reverse));
  }

  .sm\\:self-auto {
    align-self: auto;
  }

  .sm\\:rounded-lg {
    border-radius: var(--radius);
  }

  .sm\\:p-3 {
    padding: 0.75rem;
  }

  .sm\\:p-4 {
    padding: 1rem;
  }

  .sm\\:p-6 {
    padding: 1.5rem;
  }

  .sm\\:px-2 {
    padding-left: 0.5rem;
    padding-right: 0.5rem;
  }

  .sm\\:px-6 {
    padding-left: 1.5rem;
    padding-right: 1.5rem;
  }

  .sm\\:py-12 {
    padding-top: 3rem;
    padding-bottom: 3rem;
  }

  .sm\\:pb-6 {
    padding-bottom: 1.5rem;
  }

  .sm\\:pt-4 {
    padding-top: 1rem;
  }

  .sm\\:pt-6 {
    padding-top: 1.5rem;
  }

  .sm\\:text-left {
    text-align: left;
  }

  .sm\\:text-2xl {
    font-size: 1.5rem;
    line-height: 2rem;
  }

  .sm\\:text-4xl {
    font-size: 2.25rem;
    line-height: 2.5rem;
  }

  .sm\\:text-base {
    font-size: 1rem;
    line-height: 1.5rem;
  }

  .sm\\:text-lg {
    font-size: 1.125rem;
    line-height: 1.75rem;
  }

  .sm\\:text-sm {
    font-size: 0.875rem;
    line-height: 1.25rem;
  }

  .sm\\:text-xl {
    font-size: 1.25rem;
    line-height: 1.75rem;
  }

  .data-\\[state\\=open\\]\\:sm\\:slide-in-from-bottom-full[data-state="open"] {
    --tw-enter-translate-y: 100%;
  }
}

@media (min-width: 768px) {
  .md\\:absolute {
    position: absolute;
  }

  .md\\:block {
    display: block;
  }

  .md\\:flex {
    display: flex;
  }

  .md\\:w-\\[var\\(--radix-navigation-menu-viewport-width\\)\\] {
    width: var(--radix-navigation-menu-viewport-width);
  }

  .md\\:w-auto {
    width: auto;
  }

  .md\\:max-w-\\[420px\\] {
    max-width: 420px;
  }

  .md\\:grid-cols-2 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .md\\:grid-cols-3 {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .md\\:grid-cols-4 {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .md\\:text-sm {
    font-size: 0.875rem;
    line-height: 1.25rem;
  }

  .md\\:opacity-0 {
    opacity: 0;
  }

  .after\\:md\\:hidden::after {
    content: var(--tw-content);
    display: none;
  }

  .peer[data-variant="inset"] ~ .md\\:peer-data-\\[variant\\=inset\\]\\:m-2 {
    margin: 0.5rem;
  }

  .peer[data-state="collapsed"][data-variant="inset"] ~ .md\\:peer-data-\\[state\\=collapsed\\]\\:peer-data-\\[variant\\=inset\\]\\:ml-2 {
    margin-left: 0.5rem;
  }

  .peer[data-variant="inset"] ~ .md\\:peer-data-\\[variant\\=inset\\]\\:ml-0 {
    margin-left: 0px;
  }

  .peer[data-variant="inset"] ~ .md\\:peer-data-\\[variant\\=inset\\]\\:rounded-xl {
    border-radius: 0.75rem;
  }

  .peer[data-variant="inset"] ~ .md\\:peer-data-\\[variant\\=inset\\]\\:shadow {
    --tw-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1);
    --tw-shadow-colored: 0 1px 3px 0 var(--tw-shadow-color), 0 1px 2px -1px var(--tw-shadow-color);
    box-shadow: var(--tw-ring-offset-shadow, 0 0 #0000), var(--tw-ring-shadow, 0 0 #0000), var(--tw-shadow);
  }
}

@media (min-width: 1024px) {
  .lg\\:col-span-1 {
    grid-column: span 1 / span 1;
  }

  .lg\\:grid-cols-2 {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .lg\\:grid-cols-3 {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .lg\\:px-8 {
    padding-left: 2rem;
    padding-right: 2rem;
  }
}

.\\[\\&\\:has\\(\\[aria-selected\\]\\)\\]\\:bg-accent:has([aria-selected]) {
  background-color: hsl(var(--accent));
}

.first\\:\\[\\&\\:has\\(\\[aria-selected\\]\\)\\]\\:rounded-l-md:has([aria-selected]):first-child {
  border-top-left-radius: calc(var(--radius) - 2px);
  border-bottom-left-radius: calc(var(--radius) - 2px);
}

.last\\:\\[\\&\\:has\\(\\[aria-selected\\]\\)\\]\\:rounded-r-md:has([aria-selected]):last-child {
  border-top-right-radius: calc(var(--radius) - 2px);
  border-bottom-right-radius: calc(var(--radius) - 2px);
}

.\\[\\&\\:has\\(\\[aria-selected\\]\\.day-outside\\)\\]\\:bg-accent\\/50:has([aria-selected].day-outside) {
  background-color: hsl(var(--accent) / 0.5);
}

.\\[\\&\\:has\\(\\[aria-selected\\]\\.day-range-end\\)\\]\\:rounded-r-md:has([aria-selected].day-range-end) {
  border-top-right-radius: calc(var(--radius) - 2px);
  border-bottom-right-radius: calc(var(--radius) - 2px);
}

.\\[\\&\\:has\\(\\[role\\=checkbox\\]\\)\\]\\:pr-0:has([role=checkbox]) {
  padding-right: 0px;
}

.\\[\\&\\>button\\]\\:hidden>button {
  display: none;
}

.\\[\\&\\>span\\:last-child\\]\\:truncate>span:last-child {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.\\[\\&\\>span\\]\\:line-clamp-1>span {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
}

.\\[\\&\\>svg\\+div\\]\\:translate-y-\\[-3px\\]>svg+div {
  --tw-translate-y: -3px;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.\\[\\&\\>svg\\]\\:absolute>svg {
  position: absolute;
}

.\\[\\&\\>svg\\]\\:left-4>svg {
  left: 1rem;
}

.\\[\\&\\>svg\\]\\:top-4>svg {
  top: 1rem;
}

.\\[\\&\\>svg\\]\\:size-3\\.5>svg {
  width: 0.875rem;
  height: 0.875rem;
}

.\\[\\&\\>svg\\]\\:size-4>svg {
  width: 1rem;
  height: 1rem;
}

.\\[\\&\\>svg\\]\\:h-2\\.5>svg {
  height: 0.625rem;
}

.\\[\\&\\>svg\\]\\:h-3>svg {
  height: 0.75rem;
}

.\\[\\&\\>svg\\]\\:w-2\\.5>svg {
  width: 0.625rem;
}

.\\[\\&\\>svg\\]\\:w-3>svg {
  width: 0.75rem;
}

.\\[\\&\\>svg\\]\\:shrink-0>svg {
  flex-shrink: 0;
}

.\\[\\&\\>svg\\]\\:text-destructive>svg {
  color: hsl(var(--destructive));
}

.\\[\\&\\>svg\\]\\:text-foreground>svg {
  color: hsl(var(--foreground));
}

.\\[\\&\\>svg\\]\\:text-muted-foreground>svg {
  color: hsl(var(--muted-foreground));
}

.\\[\\&\\>svg\\]\\:text-sidebar-accent-foreground>svg {
  color: hsl(var(--sidebar-accent-foreground));
}

.\\[\\&\\>svg\\~\\*\\]\\:pl-7>svg~* {
  padding-left: 1.75rem;
}

.\\[\\&\\>tr\\]\\:last\\:border-b-0:last-child>tr {
  border-bottom-width: 0px;
}

.\\[\\&\\[data-panel-group-direction\\=vertical\\]\\>div\\]\\:rotate-90[data-panel-group-direction=vertical]>div {
  --tw-rotate: 90deg;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.\\[\\&\\[data-state\\=open\\]\\>svg\\]\\:rotate-180[data-state=open]>svg {
  --tw-rotate: 180deg;
  transform: translate(var(--tw-translate-x), var(--tw-translate-y)) rotate(var(--tw-rotate)) skewX(var(--tw-skew-x)) skewY(var(--tw-skew-y)) scaleX(var(--tw-scale-x)) scaleY(var(--tw-scale-y));
}

.\\[\\&_\\.recharts-cartesian-axis-tick_text\\]\\:fill-muted-foreground .recharts-cartesian-axis-tick text {
  fill: hsl(var(--muted-foreground));
}

.\\[\\&_\\.recharts-cartesian-grid_line\\[stroke\\=\\'\\#ccc\\'\\]\\]\\:stroke-border\\/50 .recharts-cartesian-grid line[stroke='#ccc'] {
  stroke: hsl(var(--border) / 0.5);
}

.\\[\\&_\\.recharts-curve\\.recharts-tooltip-cursor\\]\\:stroke-border .recharts-curve.recharts-tooltip-cursor {
  stroke: hsl(var(--border));
}

.\\[\\&_\\.recharts-dot\\[stroke\\=\\'\\#fff\\'\\]\\]\\:stroke-transparent .recharts-dot[stroke='#fff'] {
  stroke: transparent;
}

.\\[\\&_\\.recharts-layer\\]\\:outline-none .recharts-layer {
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.\\[\\&_\\.recharts-polar-grid_\\[stroke\\=\\'\\#ccc\\'\\]\\]\\:stroke-border .recharts-polar-grid [stroke='#ccc'] {
  stroke: hsl(var(--border));
}

.\\[\\&_\\.recharts-radial-bar-background-sector\\]\\:fill-muted .recharts-radial-bar-background-sector {
  fill: hsl(var(--muted));
}

.\\[\\&_\\.recharts-rectangle\\.recharts-tooltip-cursor\\]\\:fill-muted .recharts-rectangle.recharts-tooltip-cursor {
  fill: hsl(var(--muted));
}

.\\[\\&_\\.recharts-reference-line_\\[stroke\\=\\'\\#ccc\\'\\]\\]\\:stroke-border .recharts-reference-line [stroke='#ccc'] {
  stroke: hsl(var(--border));
}

.\\[\\&_\\.recharts-sector\\[stroke\\=\\'\\#fff\\'\\]\\]\\:stroke-transparent .recharts-sector[stroke='#fff'] {
  stroke: transparent;
}

.\\[\\&_\\.recharts-sector\\]\\:outline-none .recharts-sector {
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.\\[\\&_\\.recharts-surface\\]\\:outline-none .recharts-surface {
  outline: 2px solid transparent;
  outline-offset: 2px;
}

.\\[\\&_\\[cmdk-group-heading\\]\\]\\:px-2 [cmdk-group-heading] {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.\\[\\&_\\[cmdk-group-heading\\]\\]\\:py-1\\.5 [cmdk-group-heading] {
  padding-top: 0.375rem;
  padding-bottom: 0.375rem;
}

.\\[\\&_\\[cmdk-group-heading\\]\\]\\:text-xs [cmdk-group-heading] {
  font-size: 0.75rem;
  line-height: 1rem;
}

.\\[\\&_\\[cmdk-group-heading\\]\\]\\:font-medium [cmdk-group-heading] {
  font-weight: 500;
}

.\\[\\&_\\[cmdk-group-heading\\]\\]\\:text-muted-foreground [cmdk-group-heading] {
  color: hsl(var(--muted-foreground));
}

.\\[\\&_\\[cmdk-group\\]\\:not\\(\\[hidden\\]\\)_\\~\\[cmdk-group\\]\\]\\:pt-0 [cmdk-group]:not([hidden]) ~[cmdk-group] {
  padding-top: 0px;
}

.\\[\\&_\\[cmdk-group\\]\\]\\:px-2 [cmdk-group] {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.\\[\\&_\\[cmdk-input-wrapper\\]_svg\\]\\:h-5 [cmdk-input-wrapper] svg {
  height: 1.25rem;
}

.\\[\\&_\\[cmdk-input-wrapper\\]_svg\\]\\:w-5 [cmdk-input-wrapper] svg {
  width: 1.25rem;
}

.\\[\\&_\\[cmdk-input\\]\\]\\:h-12 [cmdk-input] {
  height: 3rem;
}

.\\[\\&_\\[cmdk-item\\]\\]\\:px-2 [cmdk-item] {
  padding-left: 0.5rem;
  padding-right: 0.5rem;
}

.\\[\\&_\\[cmdk-item\\]\\]\\:py-3 [cmdk-item] {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
}

.\\[\\&_\\[cmdk-item\\]_svg\\]\\:h-5 [cmdk-item] svg {
  height: 1.25rem;
}

.\\[\\&_\\[cmdk-item\\]_svg\\]\\:w-5 [cmdk-item] svg {
  width: 1.25rem;
}

.\\[\\&_p\\]\\:leading-relaxed p {
  line-height: 1.625;
}

.\\[\\&_svg\\]\\:pointer-events-none svg {
  pointer-events: none;
}

.\\[\\&_svg\\]\\:size-4 svg {
  width: 1rem;
  height: 1rem;
}

.\\[\\&_svg\\]\\:shrink-0 svg {
  flex-shrink: 0;
}

.\\[\\&_tr\\:last-child\\]\\:border-0 tr:last-child {
  border-width: 0px;
}

.\\[\\&_tr\\]\\:border-b tr {
  border-bottom-width: 1px;
}

[data-side=left][data-collapsible=offcanvas] .\\[\\[data-side\\=left\\]\\[data-collapsible\\=offcanvas\\]_\\&\\]\\:-right-2 {
  right: -0.5rem;
}

[data-side=left][data-state=collapsed] .\\[\\[data-side\\=left\\]\\[data-state\\=collapsed\\]_\\&\\]\\:cursor-e-resize {
  cursor: e-resize;
}

[data-side=left] .\\[\\[data-side\\=left\\]_\\&\\]\\:cursor-w-resize {
  cursor: w-resize;
}

[data-side=right][data-collapsible=offcanvas] .\\[\\[data-side\\=right\\]\\[data-collapsible\\=offcanvas\\]_\\&\\]\\:-left-2 {
  left: -0.5rem;
}

[data-side=right][data-state=collapsed] .\\[\\[data-side\\=right\\]\\[data-state\\=collapsed\\]_\\&\\]\\:cursor-w-resize {
  cursor: w-resize;
}

[data-side=right] .\\[\\[data-side\\=right\\]_\\&\\]\\:cursor-e-resize {
  cursor: e-resize;
}
`;window.__BENCHMARK_CSS__=Oie;function Vg(){try{if(typeof window>"u")return;const e=document.querySelectorAll('[data-react-component="benchmark-dashboard"]');if(e.length===0)return;e.forEach((t,n)=>{var r;try{const a=t;if(a.dataset.reactInitialized==="true")return;const o={apiBase:a.dataset.apiBase,initialVersion:a.dataset.version,theme:a.dataset.theme||"light",containerClassName:a.dataset.containerClass||"",containerId:`benchmark-dashboard-${n}`,features:{header:a.dataset.showHeader==="true",versionSelector:a.dataset.showVersionSelector!=="false",summaryCards:a.dataset.showSummaryCards!=="false",tabs:((r=a.dataset.tabs)==null?void 0:r.split(",").map(s=>s.trim()))||["overview","latency","resources"]}};Ic.createRoot(a).render(b.jsx(E.StrictMode,{children:b.jsx(Sie,{...o})})),a.dataset.reactInitialized="true"}catch(a){console.error(`Error initializing benchmark dashboard ${n+1}:`,a)}})}catch(e){console.error("Error in benchmark dashboard initialization:",e)}}function N2(){setTimeout(()=>{Vg()},100)}document.readyState==="loading"?document.addEventListener("DOMContentLoaded",N2):N2();window.addEventListener("load",()=>{setTimeout(Vg,200)});window.initializeBenchmarkDashboardShadow=Vg;
