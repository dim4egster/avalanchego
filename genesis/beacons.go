// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"github.com/dim4egster/avalanchego/utils/constants"
	"github.com/dim4egster/avalanchego/utils/sampler"
)

// getIPs returns the beacon IPs for each network
func getIPs(networkID uint32) []string {
	switch networkID {
	case constants.MainnetID:
		return []string{
			//XXX.XXX.XXX.XXX:PORT",
		}
	case constants.FujiID:
		return []string{
			//XXX.XXX.XXX.XXX:PORT",
		}
	default:
		return nil
	}
}

// getNodeIDs returns the beacon node IDs for each network
func getNodeIDs(networkID uint32) []string {
	switch networkID {
	case constants.MainnetID:
		return []string{
			"NodeID-6ZBhT4m9kNQ4cpeDoKSb1PBftYAg2Ltfx",
			"NodeID-7fsWLv7iCMEbobu2AjdwCbnwMtM2ahSWW",
			"NodeID-LPDaqcLiC77vdHz6TDsjUS4T6DYAEFhaB",
			"NodeID-HKR2zonLiAGt2PqeRzfyLSrsJyEk9HVDV",
			"NodeID-JJKqoEaA8KStxJYm8GZ8qexz27kbya9Fm",
		}
	case constants.FujiID:
		return []string{
			//NodeID's for tesntet
		}
	default:
		return nil
	}
}

// SampleBeacons returns the some beacons this node should connect to
func SampleBeacons(networkID uint32, count int) ([]string, []string) {
	ips := getIPs(networkID)
	ids := getNodeIDs(networkID)

	if numIPs := len(ips); numIPs < count {
		count = numIPs
	}

	sampledIPs := make([]string, 0, count)
	sampledIDs := make([]string, 0, count)

	s := sampler.NewUniform()
	_ = s.Initialize(uint64(len(ips)))
	indices, _ := s.Sample(count)
	for _, index := range indices {
		sampledIPs = append(sampledIPs, ips[int(index)])
		sampledIDs = append(sampledIDs, ids[int(index)])
	}

	return sampledIPs, sampledIDs
}
