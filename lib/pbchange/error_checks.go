package pbchange

/***********************************************************************
   Copyright 2018 Information Trust Institute

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
***********************************************************************/

func IsCMError(e error) bool {
	switch e.(type) {
	case CMError:
		return true
	}
	return false
}

func IsCMInitError(e error) bool {
	switch e.(type) {
	case CMInitError:
		return true
	}
	return false
}

func IsCMCommunicationError(e error) bool {
	switch e.(type) {
	case CMCommunicationError:
		return true
	}
	return false
}

func IsCMNoRepoError(e error) bool {
	switch e.(type) {
	case CMNoRepoError:
		return true
	}
	return false
}

func IsCMMetaNoKeyError(e error) bool {
	switch e.(type) {
	case CMMetaNoKeyError:
		return true
	}
	return false
}

func IsCMCollisionError(e error) bool {
	switch e.(type) {
	case CMCollisionError:
		return true
	}
	return false
}

func IsCMTransactionError(e error) bool {
	switch e.(type) {
	case CMTransactionError:
		return true
	}
	return false
}

func IsCMIDError(e error) bool {
	switch e.(type) {
	case CMIDError:
		return true
	}
	return false
}
