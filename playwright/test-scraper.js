// Simple test script for the scraper service
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

const PROTO_PATH = '../proto/librus_scraper.proto';

// Load protobuf
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const protoDescriptor = grpc.loadPackageDefinition(packageDefinition);
const librusScraperProto = protoDescriptor.librus_scraper;

// Create client
const client = new librusScraperProto.LibrusScraper(
  'localhost:50051',
  grpc.credentials.createInsecure()
);

async function testHealthCheck() {
  return new Promise((resolve, reject) => {
    client.healthCheck({}, (error, response) => {
      if (error) {
        reject(error);
      } else {
        resolve(response);
      }
    });
  });
}

async function testGetAllUpdates() {
  return new Promise((resolve, reject) => {
    client.getAllUpdates({
      login: '10711687',
      password: 'Smlrsdf123'
    }, (error, response) => {
      if (error) {
        reject(error);
      } else {
        resolve(response);
      }
    });
  });
}

async function runTests() {
  try {
    console.log('Testing health check...');
    const health = await testHealthCheck();
    console.log('Health check result:', health);
    
    console.log('\nTesting get all updates...');
    const updates = await testGetAllUpdates();
    console.log('Updates result:', {
      success: updates.success,
      messagesCount: updates.messages?.length || 0,
      newsCount: updates.news?.length || 0,
      error: updates.error_message
    });
    
    if (updates.messages?.length > 0) {
      console.log('\nFirst message:', {
        id: updates.messages[0].id,
        title: updates.messages[0].title,
        author: updates.messages[0].author
      });
    }
    
  } catch (error) {
    console.error('Test failed:', error);
  }
}

runTests();
