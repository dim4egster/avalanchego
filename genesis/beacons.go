// Copyright (C) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"github.com/dim4egster/qmallgo/utils/constants"
	"github.com/dim4egster/qmallgo/utils/sampler"
)

// getIPs returns the beacon IPs for each network
func getIPs(networkID uint32) []string {
	switch networkID {
	case constants.MainnetID:
		return []string{
						"127.0.0.1:9671",
						"127.0.0.1:9673",
						"127.0.0.1:9675",
						"127.0.0.1:9677",
						"127.0.0.1:9679",
						/*"127.0.0.1:9681",
						"127.0.0.1:9683",
						"127.0.0.1:9685",
						"127.0.0.1:9687",
						"127.0.0.1:9689",*/
		}
	case constants.FujiID:
		return []string{
			"127.0.0.1:9651",
			"127.0.0.1:9653",
			"127.0.0.1:9655",
			"127.0.0.1:9657",
			"127.0.0.1:9659",
			/*"127.0.0.1:9661",
			"127.0.0.1:9663",
			"127.0.0.1:9665",
			"127.0.0.1:9667",
			"127.0.0.1:9669",*/
		}
	default:
		return nil
	}
}

// getNodeIDs returns the beacon node IDs for each network
func getNodeIDs(networkID uint32) []string {
	switch networkID {
	case constants.MainnetID:
		return []string {
			"NodeID-2YXYFUDkEf8posCKwKfdVjT7y571fNLQs",
			"NodeID-CwpnhaPUX3T1PDqm45kdrep6hmv4kS9LH",
			"NodeID-QKuMoE1VJJYRrdprWSgM9iAu6jjUcfcET",
			"NodeID-8RioHCVCoU1aqgt7HxCQpDPkRPysh8wNS",
			"NodeID-6tukgiieXaEaAF7bVW6rgzuXbQMc6m2VM",
			/*"NodeID-D8zWh96fqq4jTrfnCrMratNrdQ9B6VZVN",
			"NodeID-Mo3RTo1TLeT1N3R1B4M56CGEdgkyrbK4a",
			"NodeID-BeNKtHJLU858bYCPDPhtkckdeggVcvHDm",
			"NodeID-LZk4m1pi2cTdbJZ5xpo8kHQ337H1P2xc3",
			"NodeID-LX9kbRaSoFcSmf7tx1fpukytf518cJuAT",*/
		}
	case constants.FujiID:
		return []string{
			"NodeID-6ZBhT4m9kNQ4cpeDoKSb1PBftYAg2Ltfx",
			"NodeID-7fsWLv7iCMEbobu2AjdwCbnwMtM2ahSWW",
			"NodeID-LPDaqcLiC77vdHz6TDsjUS4T6DYAEFhaB",
			"NodeID-HKR2zonLiAGt2PqeRzfyLSrsJyEk9HVDV",
			"NodeID-JJKqoEaA8KStxJYm8GZ8qexz27kbya9Fm",
			/*"NodeID-MCWFX8GuYBjZ34YunxfaigxdUQ8sCaZHC",
			"NodeID-RLUDk9bLSexV2moHmr45HCbH6LrNGcNx",
			"NodeID-H1LYqkUPu3ZZSbSdwv5vpsUrj7wyjPPvk",
			"NodeID-FDoKRfcHAve2bHt2tkKcDNsYJuvqQf9tT",
			"NodeID-GfQjYRGzgxr1saZqWPdQZWB5unFzBFz7R",*/
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
