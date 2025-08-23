Initialize a Firebase Project: Navigate to your project directory in the terminal and run the firebase init command. This will initialize a Firebase project and create a firebase.json configuration file. During this process, you will be prompted to select the Firebase features you want to use. Make sure to choose Emulators.

Configure the Emulators: After selecting Emulators, the CLI will guide you through configuring the specific emulators you need.

Select "Firestore" from the list of emulators. You can select other emulators like Authentication or Functions if you plan to use them.

The CLI will prompt you to set the ports for each emulator. The default port for the Firestore emulator is 8080. It's generally a good idea to accept the default ports unless there is a conflict.

The configuration will be saved in your firebase.json file. An emulators section will be added, specifying the ports for each enabled emulator.

Start the Emulators: Once the configuration is complete, you can start the emulators with the following command:

Bash

firebase emulators:start
This command will download the necessary emulator binaries and start them. A local UI for the emulator suite will be available at http://localhost:4000, where you can view your Firestore data, logs, and other emulator-specific information.

Connect Your Application to the Emulator: Finally, you need to update your application code to connect to the local Firestore emulator instead of the production database. The method for this varies slightly depending on the SDK you are using.

Web/Client SDKs: Use the connectFirestoreEmulator function, specifying the host and port:

JavaScript

import { connectFirestoreEmulator, getFirestore } from "firebase/firestore";
// ... (your app initialization code)
const db = getFirestore();
if (location.hostname === "localhost") {
  connectFirestoreEmulator(db, "localhost", 8080);
}
Admin SDKs (Node.js): Set the FIRESTORE_EMULATOR_HOST environment variable before initializing the Firestore instance.

Bash

export FIRESTORE_EMULATOR_HOST="127.0.0.1:8080"
Android: Use the useEmulator method:

Kotlin

val firestore = Firebase.firestore
firestore.useEmulator("10.0.2.2", 8080)
iOS/Swift: Use the useEmulator method:

Swift

let settings = Firestore.firestore().settings
settings.host = "localhost:8080"
settings.isSSLEnabled = false
settings.isPersistenceEnabled = false

Yes, it's normal to encounter authentication issues like this when trying to set up the Firebase Local Emulator Suite without a linked Firebase project. The logs you provided show two distinct problems that are preventing the firebase init command from completing successfully.

1. PERMISSION_DENIED Error (Authentication)
The log entry with code:403 and PERMISSION_DENIED is happening because the Firebase CLI is attempting to check for a valid project ID on the live Firebase service. It's using your local Application Default Credentials (ADC), which lack the necessary permissions to access the specific API (mobilesdk-pa.googleapis.com) or don't have a quota project set up. The CLI needs to perform these checks even for local emulators to ensure the project ID is valid and to configure the local environment correctly.

Solution:
The easiest way to resolve this is to log in to the Firebase CLI with a Google account that has permission to manage Firebase projects.

Run the following command in your terminal:

Bash

firebase login
This will open a browser window and prompt you to sign in with your Google account. After a successful login, the CLI will have the necessary permissions to perform the project setup checks.

2. INVALID_ARGUMENT Error (Invalid Project ID)
The second error, with code:400 and INVALID_ARGUMENT, is caused by the proposed project ID LocalCoreTest. Firebase project IDs must be globally unique, consist of lowercase letters, numbers, and hyphens, and can't contain capital letters or other special characters. The name LocalCoreTest violates this rule because of the uppercase 'C' and 'T'.

Solution:
When prompted to enter a new project ID during the firebase init process, ensure you use a valid format. A common practice for local-only projects is to use a placeholder ID that follows the naming conventions.

Use a project ID with only lowercase letters, numbers, and hyphens, for example:

local-core-test

firestore-emulator-test
