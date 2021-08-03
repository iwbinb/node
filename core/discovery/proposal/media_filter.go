/*
 * Copyright (C) 2021 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package proposal

import (
	"sort"

	"github.com/mysteriumnetwork/node/market"
)

func mediaStreamingFilter(proposals []market.ServiceProposal) []market.ServiceProposal {
	var totalBandwidth, averageBandwidth float64
	var totalQuality, avgQuality float64

	for _, p := range proposals {
		totalBandwidth += p.Quality.Bandwidth
		totalQuality += p.Quality.Quality
	}
	averageBandwidth = totalBandwidth / float64(len(proposals))
	avgQuality = avgQualityCeiling(totalQuality / float64(len(proposals)))

	var filtered []market.ServiceProposal
	for _, p := range proposals {
		if p.Quality.Quality >= avgQuality && p.Quality.Bandwidth >= averageBandwidth && p.Location.IPType == "residential" {
			filtered = append(filtered, p)
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		qx, qy := filtered[i].Quality, filtered[j].Quality

		if qx.Bandwidth == qy.Bandwidth {
			return qx.Quality > qy.Quality
		}
		return qx.Bandwidth > qy.Bandwidth
	})

	return filtered
}
