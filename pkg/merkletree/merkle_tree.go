package merkletree

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

var Trees []MerkleTree

type MerkleTree struct {
	ID   uuid.UUID
	Root *Node
}

type Node struct {
	Hash  []byte
	Left  *Node
	Right *Node
}

type MerkleProof []ProofSibling

type ProofSibling struct {
	Hash   []byte
	IsLeft bool
}

func BuildTree(hashes [][]byte) *Node {
	var currentLevel []*Node

	for _, hash := range hashes {
		currentLevel = append(currentLevel, newNode(hash, nil, nil))
	}

	// If len(currentLevel) is 1 we are at the root
	for len(currentLevel) > 1 {
		var nextLevel []*Node

		// Is a binary tree so each node must have two children
		if len(currentLevel)%2 != 0 {
			currentLevel = append(currentLevel, currentLevel[len(currentLevel)-1])
		}

		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := currentLevel[i+1]
			hash := hashPair(left.Hash, right.Hash)

			newNode := newNode(hash, left, right)
			nextLevel = append(nextLevel, newNode)
		}

		currentLevel = nextLevel
	}

	return currentLevel[0]
}

func AddTree(newTree MerkleTree) {
	Trees = append(Trees, newTree)
}

func GetTree(treeID uuid.UUID) *MerkleTree {
	for _, tree := range Trees {
		if tree.ID == treeID {
			return &tree
		}
	}
	return nil
}

func UpdateTree(newTree MerkleTree) {
	for i, tree := range Trees {
		if tree.ID == newTree.ID {
			Trees[i] = newTree
			return
		}
	}

	AddTree(newTree)
}

func CreateMerkleProof(root *Node, hash []byte) (MerkleProof, error) {
	var proof MerkleProof

	var findHash func(node *Node) bool
	findHash = func(node *Node) bool {
		if node == nil {
			return false
		}

		if node.Left == nil && node.Right == nil {
			return bytes.Equal(node.Hash, hash)
		}

		leftContainsHash := findHash(node.Left)
		if leftContainsHash {
			sibling := ProofSibling{
				Hash:   node.Right.Hash,
				IsLeft: false,
			}
			proof = append(proof, sibling)
			return true
		}

		rightContainsHash := findHash(node.Right)
		if rightContainsHash {
			sibling := ProofSibling{
				Hash:   node.Left.Hash,
				IsLeft: true,
			}
			proof = append(proof, sibling)
			return true
		}

		return false
	}

	hashFound := findHash(root)
	if !hashFound {
		return nil, fmt.Errorf("hash not found in the Merkle tree")
	}

	return proof, nil
}

func VerifyMerkleProof(rootHash []byte, hash []byte, proof MerkleProof) bool {
	currentHash := hash
	for _, sibling := range proof {
		if sibling.IsLeft {
			currentHash = hashPair(sibling.Hash, currentHash)
		} else {
			currentHash = hashPair(currentHash, sibling.Hash)
		}
	}

	return bytes.Equal(currentHash, rootHash)
}

func newNode(hash []byte, left *Node, right *Node) *Node {
	return &Node{
		Hash:  hash,
		Left:  left,
		Right: right,
	}
}

func hashPair(left []byte, right []byte) []byte {
	pair := append(left, right...)
	hash := sha256.Sum256(pair)
	return hash[:]
}
