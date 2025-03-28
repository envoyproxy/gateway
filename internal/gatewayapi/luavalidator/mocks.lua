-- ParsedName object.
local ParsedName = {}

-- Logging methods.
function ParsedName:logTrace(message)
  print("[TRACE]: " .. message)
end

function ParsedName:logDebug(message)
  print("[DEBUG]: " .. message)
end

function ParsedName:logInfo(message)
  print("[INFO]: " .. message)
end

function ParsedName:logWarn(message)
  print("[WARN]: " .. message)
end

function ParsedName:logErr(message)
  print("[ERROR]: " .. message)
end

function ParsedName:logCritical(message)
  print("[CRITICAL]: " .. message)
end

-- Method to return the common name.
function ParsedName:commonName()
  return "common_name"
end

-- Method to return the organization name.
function ParsedName:organizationName()
  return "organization_name"
end

local SSLConnectionObject = {}

-- Logging methods
function SSLConnectionObject:logTrace(message)
  print("[TRACE]: " .. message)
end

function SSLConnectionObject:logDebug(message)
  print("[DEBUG]: " .. message)
end

function SSLConnectionObject:logInfo(message)
  print("[INFO]: " .. message)
end

function SSLConnectionObject:logWarn(message)
  print("[WARN]: " .. message)
end

function SSLConnectionObject:logErr(message)
  print("[ERROR]: " .. message)
end

function SSLConnectionObject:logCritical(message)
  print("[CRITICAL]: " .. message)
end

-- SSL-related methods
function SSLConnectionObject:peerCertificatePresented()
  return true -- Simulate that the peer certificate is always presented
end

function SSLConnectionObject:peerCertificateValidated()
  return false -- Simulate no validation for TLS session resumption
end

function SSLConnectionObject:uriSanLocalCertificate()
  local t = {}
  for i = 1, 1000 do
    t[i] = "san"
  end
  return t
end

function SSLConnectionObject:sha256PeerCertificateDigest()
  return "abcdef1234567890" -- Simulated SHA256 digest
end

function SSLConnectionObject:serialNumberPeerCertificate()
  return "123456789" -- Simulated serial number
end

function SSLConnectionObject:issuerPeerCertificate()
  return "CN=Mock Issuer, O=Example Org" -- Simulated issuer
end

function SSLConnectionObject:subjectPeerCertificate()
  return "CN=Mock Subject, O=Example Org" -- Simulated subject
end

function SSLConnectionObject:parsedSubjectPeerCertificate()
  return {
    commonName = function() return "Mock Subject" end,
    organizationName = function() return "Example Org" end,
  }
end

function SSLConnectionObject:uriSanPeerCertificate()
  return {"uri1", "uri2"} -- Example SAN URIs for peer certificate
end

function SSLConnectionObject:subjectLocalCertificate()
  return "CN=Local Cert, O=Example Org" -- Simulated local certificate subject
end

function SSLConnectionObject:urlEncodedPemEncodedPeerCertificate()
  return "PEM_ENCODED_PEER_CERT" -- Simulated PEM encoding
end

function SSLConnectionObject:urlEncodedPemEncodedPeerCertificateChain()
  return "PEM_ENCODED_CERT_CHAIN" -- Simulated PEM chain
end

function SSLConnectionObject:dnsSansPeerCertificate()
  local t = {}
    for i = 1, 1000 do
      t[i] = "example.com"
    end
  return t
end

function SSLConnectionObject:dnsSansLocalCertificate()
  local t = {}
    for i = 1, 1000 do
      t[i] = "local.com"
    end
  return t
end

function SSLConnectionObject:oidsPeerCertificate()
  local t = {}
    for i = 1, 1000 do
      t[i] = "1.2.840.113549.1.1.1"
    end
  return t
end

function SSLConnectionObject:oidsLocalCertificate()
  local t = {}
    for i = 1, 1000 do
      t[i] = "1.2.840.113549.1.1.1"
    end
  return t
end

function SSLConnectionObject:validFromPeerCertificate()
  return 1672531200
end

function SSLConnectionObject:expirationPeerCertificate()
  return 2724608000
end

function SSLConnectionObject:sessionId()
  return "abcdef1234567890abcdef1234567890"
end

function SSLConnectionObject:ciphersuiteId()
  return "0x1301"
end

function SSLConnectionObject:ciphersuiteString()
  return "TLS_AES_128_GCM_SHA256"
end

function SSLConnectionObject:tlsVersion()
  return "TLSv1.3"
end

-- Connection object
local Connection = {}

-- Logging methods for Connection
function Connection:logTrace(message)
  print("[TRACE]: " .. message)
end

function Connection:logDebug(message)
  print("[DEBUG]: " .. message)
end

function Connection:logInfo(message)
  print("[INFO]: " .. message)
end

function Connection:logWarn(message)
  print("[WARN]: " .. message)
end

function Connection:logErr(message)
  print("[ERROR]: " .. message)
end

function Connection:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function Connection:ssl()
  return SSLConnectionObject
end

local DynamicMetadata = {
  data = {
    key1 = "value1",
    key2 = "value2"
  }
}

setmetatable(DynamicMetadata, {
  __pairs = function(self)
    return pairs(self.data)
  end,
  __index = function(self, key)
    return self.data[key]
  end
})

-- Logging methods
function DynamicMetadata:logTrace(message)
  print("[TRACE]: " .. message)
end

function DynamicMetadata:logDebug(message)
  print("[DEBUG]: " .. message)
end

function DynamicMetadata:logInfo(message)
  print("[INFO]: " .. message)
end

function DynamicMetadata:logWarn(message)
  print("[WARN]: " .. message)
end

function DynamicMetadata:logErr(message)
  print("[ERROR]: " .. message)
end

function DynamicMetadata:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function DynamicMetadata:get(filterName)
  return "get_result"
end

function DynamicMetadata:set(filterName, key, value)
end

-- ConnectionStreamInfo Object
local ConnectionStreamInfo = {}

-- Logging methods for ConnectionStreamInfo
function ConnectionStreamInfo:logTrace(message)
  print("[TRACE]: " .. message)
end

function ConnectionStreamInfo:logDebug(message)
  print("[DEBUG]: " .. message)
end

function ConnectionStreamInfo:logInfo(message)
  print("[INFO]: " .. message)
end

function ConnectionStreamInfo:logWarn(message)
  print("[WARN]: " .. message)
end

function ConnectionStreamInfo:logErr(message)
  print("[ERROR]: " .. message)
end

function ConnectionStreamInfo:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function ConnectionStreamInfo:dynamicMetadata()
  return DynamicMetadata
end

-- StreamInfo Object
local StreamInfo = {}

function StreamInfo:logTrace(message)
  print("[TRACE]: " .. message)
end

function StreamInfo:logDebug(message)
  print("[DEBUG]: " .. message)
end

function StreamInfo:logInfo(message)
  print("[INFO]: " .. message)
end

function StreamInfo:logWarn(message)
  print("[WARN]: " .. message)
end

function StreamInfo:logErr(message)
  print("[ERROR]: " .. message)
end

function StreamInfo:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function StreamInfo:protocol()
  return "HTTP/2"
end

function StreamInfo:routeName()
  return "example-route"
end

function StreamInfo:virtualClusterName()
  return "example-virtual-cluster"
end

function StreamInfo:downstreamDirectLocalAddress()
  return "127.0.0.1:8080"
end

function StreamInfo:downstreamLocalAddress()
  return "192.168.1.100:8080"
end

function StreamInfo:downstreamDirectRemoteAddress()
  return "192.168.1.50:5050"
end

function StreamInfo:downstreamRemoteAddress()
  return "10.0.0.1:5050"
end

function StreamInfo:dynamicMetadata()
  return DynamicMetadata
end

function StreamInfo:downstreamSslConnection()
  return SSLConnectionObject
end

function StreamInfo:requestedServerName()
  return "example-server"
end

-- Metadata Object
local Metadata = {
  data = {
    key1 = "value1",
    key2 = "value2"
  }
}

setmetatable(Metadata, {
  __pairs = function(self)
    return pairs(self.data)
  end,
  __index = function(self, key)
    return self.data[key]
  end
})

-- Logging Methods
function Metadata:logTrace(message)
  print("[TRACE]: " .. message)
end

function Metadata:logDebug(message)
  print("[DEBUG]: " .. message)
end

function Metadata:logInfo(message)
  print("[INFO]: " .. message)
end

function Metadata:logWarn(message)
  print("[WARN]: " .. message)
end

function Metadata:logErr(message)
  print("[ERROR]: " .. message)
end

function Metadata:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function Metadata:get(key)
  return "get_example"
end

-- Buffer Object
local Buffer = {}

-- Logging Methods
function Buffer:logTrace(message)
  print("[TRACE]: " .. message)
end

function Buffer:logDebug(message)
  print("[DEBUG]: " .. message)
end

function Buffer:logInfo(message)
  print("[INFO]: " .. message)
end

function Buffer:logWarn(message)
  print("[WARN]: " .. message)
end

function Buffer:logErr(message)
  print("[ERROR]: " .. message)
end

function Buffer:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function Buffer:length()
  return 10
end

function Buffer:getBytes(index, length)
  return "example_bytes"
end

function Buffer:setBytes(setTo)
end

-- Headers Object
local Headers = {
  data = {
    key1 = "value1",
    key2 = "value2"
  }
}

setmetatable(Headers, {
  __pairs = function(self)
    return pairs(self.data)
  end,
  __index = function(self, key)
    return self.data[key]
  end
})

-- Logging Methods
function Headers:logTrace(message)
  print("[TRACE]: " .. message)
end

function Headers:logDebug(message)
  print("[DEBUG]: " .. message)
end

function Headers:logInfo(message)
  print("[INFO]: " .. message)
end

function Headers:logWarn(message)
  print("[WARN]: " .. message)
end

function Headers:logErr(message)
  print("[ERROR]: " .. message)
end

function Headers:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function Headers:add(key, value)
end

function Headers:get(key)
  return "example_value"
end

function Headers:getAtIndex(key, index)
  return "example_value"
end

function Headers:getNumValues(key)
  return 1
end

function Headers:remove(key)
end

function Headers:replace(key, value)
end

function Headers:setHttp1ReasonPhrase(reasonPhrase)
end

function Headers:getHttp1ReasonPhrase()
  return "reason"
end

-- StreamHandle Object
local StreamHandle = {
data = {Buffer}
}

-- Logging Methods (with internal logging mechanism)
function StreamHandle:logTrace(message)
  print("[TRACE]: " .. message)
end

function StreamHandle:logDebug(message)
  print("[DEBUG]: " .. message)
end

function StreamHandle:logInfo(message)
  print("[INFO]: " .. message)
end

function StreamHandle:logWarn(message)
  print("[WARN]: " .. message)
end

function StreamHandle:logErr(message)
  print("[ERROR]: " .. message)
end

function StreamHandle:logCritical(message)
  print("[CRITICAL]: " .. message)
end

function StreamHandle:headers()
  return Headers
end

function StreamHandle:body(always_wrap_body)
  return "body"
end

function StreamHandle:bodyChunks()
  local index = 0
  local chunks = self.data or {}

  return function()
    index = index + 1
    local chunk = chunks[index]
    if chunk then
      return chunk
    end
  end
end

function StreamHandle:trailers()
  return Headers
end

function StreamHandle:httpCall(cluster, headers, body, timeout_ms, asynchronous)
  self:logInfo("Making HTTP call to cluster: " .. cluster)
  return headers, body
end

function StreamHandle:respond(headers, body)
  self:logInfo("Responding with status: " .. headers[":status"])
end

function StreamHandle:metadata()
  return Metadata
end

function StreamHandle:streamInfo()
  return StreamInfo
end

function StreamHandle:connection()
  return Connection
end

function StreamHandle:connectionStreamInfo()
  return ConnectionStreamInfo
end

function StreamHandle:setUpstreamOverrideHost(host, strict)
end

function StreamHandle:importPublicKey(keyder, keyderLength)
  return "mocked_public_key"
end

function StreamHandle:verifySignature(hashFunction, pubkey, signature, signatureLength, data, dataLength)
  return true, ""
end

function StreamHandle:base64Escape(inputString)
  return inputString
end

function StreamHandle:timestamp(format)
  return os.time() * 1000
end

function StreamHandle:timestampString(resolution)
  return os.date("!%Y-%m-%dT%H:%M:%SZ", os.time())
end